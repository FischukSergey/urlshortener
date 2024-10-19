package httpserver

import (
	"context"
	stdLog "log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/batch"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/deletedflag"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/getpingdb"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/geturl"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/getuserallurl"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/saveurl"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/saveurljson"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/stats"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/gzipper"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/mwlogger"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/trustsubnet"
	"github.com/FischukSergey/urlshortener.git/internal/storage"
	"github.com/FischukSergey/urlshortener.git/internal/storage/dbstorage"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
)

// NewHTTPServer инициализирует сервер, маршруты и middleware
func StartHTTPServer(log *slog.Logger) {
	//инициализация хранилища
	repo, err := storage.InitStorage(log)
	if err != nil {
		stdLog.Fatal("Ошибка инициализации хранилища", err.Error())
	}
	var storage config.URLStorage
	var delChan chan config.DeletedRequest
	switch {
	case config.FlagDatabaseDSN != "":
		storage = repo.(*dbstorage.Storage)
		delChan = repo.(*dbstorage.Storage).DelChan
	default:
		storage = repo.(*mapstorage.DataStore)
		delChan = repo.(*mapstorage.DataStore).DelChan
	}

	//инициализация роутера и middleware
	r := chi.NewRouter()             //инициализируем роутер и middleware
	r.Use(mwlogger.NewMwLogger(log)) //маршрут в middleware за логированием
	r.Use(gzipper.NewMwGzipper(log)) //работа со сжатыми запросами/сжатие ответов
	r.Use(auth.NewMwToken(log))      //ID session аутентификация пользователя/JWToken в  cookie
	r.Use(middleware.RequestID)
	r.Mount("/debug", middleware.Profiler())

	//инициализация обработчиков
	r.Get("/{alias}", geturl.GetURL(log, storage))
	r.Get("/ping", getpingdb.GetPingDB(log, storage))
	r.Get("/api/user/urls", getuserallurl.GetUserAllURL(log, storage))
	r.Post("/", saveurl.PostURL(log, storage))
	r.Post("/api/shorten", saveurljson.PostURLjson(log, storage))
	r.Post("/api/shorten/batch", batch.PostBatch(log, storage))
	r.Delete("/api/user/urls", deletedflag.DeleteShortURL(log, delChan))
	r.Group(func(r chi.Router) {
		r.Use(trustsubnet.MwTrustSubnet(log, config.FlagTrustedSubnets))
		r.Get("/api/internal/stats", stats.GetStats(log, storage))
	})

	//запускаем сервер
	srv := &http.Server{
		Addr:         config.FlagServerPort,
		Handler:      r,
		ReadTimeout:  40 * time.Second,
		WriteTimeout: 40 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Info("Initializing server", slog.String("address", srv.Addr))

	serverCtx, serverCancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig //ожидаем сигнал прерывания
		log.Info("Получен сигнал прерывания")

		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 45*time.Second) //таймаут на завершение работы сервера
		defer shutdownCancel()
		go func() { //принудительное завершение работы сервера по таймауту
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				stdLog.Fatal("таймаут при завершении работы сервера", shutdownCtx.Err())
			}
		}()

		//завершаем работу сервера
		if err := srv.Shutdown(shutdownCtx); err != nil {
			stdLog.Fatal("Ошибка при завершении работы сервера", err.Error())
		}
		serverCancel() //посылаем сигнал завершения работы для других процессов
	}()

	//запускаем сервер
	//если есть флаг TLS, то генерируем сертификат и запускаем TLS сервер
	if config.ServerTLS {
		log.Info("Starting TLS server")
		err := config.GenerateTLS() //генерируем сертификат и ключ
		if err != nil {
			stdLog.Fatal("Ошибка при генерации TLS конфигурации", err.Error())
		}
		if err := srv.ListenAndServeTLS(config.ServerCertPath, config.ServerKeyPath); err != nil {
			stdLog.Fatal("Ошибка при запуске TLS сервера", err.Error())
		}
	} else {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			stdLog.Fatal("Ошибка при запуске сервера", err.Error())
		}
	}

	<-serverCtx.Done() //ожидаем завершения работы сервера
	log.Info("Сервер завершил работу")
	// закрываем соединение с хранилищем
	storage.Close()
}
