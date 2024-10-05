package main

import (
	"context"
	"fmt"
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
	"github.com/jackc/pgx/v5/pgconn"

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
	"github.com/FischukSergey/urlshortener.git/internal/models"
	"github.com/FischukSergey/urlshortener.git/internal/storage/dbstorage"
	"github.com/FischukSergey/urlshortener.git/internal/storage/jsonstorage"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Println(
		"Build version: ", buildVersion,
		"\nBuild date: ", buildDate,
		"\nBuild commit: ", buildCommit,
	)
	var log = slog.New( //инициализируем логгер
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	config.ParseFlags() //инициализируем флаги/переменные окружения конфигурации сервера

	r := chi.NewRouter()             //инициализируем роутер и middleware
	r.Use(mwlogger.NewMwLogger(log)) //маршрут в middleware за логированием
	r.Use(gzipper.NewMwGzipper(log)) //работа со сжатыми запросами/сжатие ответов
	r.Use(auth.NewMwToken(log))      //ID session аутентификация пользователя/JWToken в  cookie
	r.Use(middleware.RequestID)
	r.Mount("/debug", middleware.Profiler())

	var mapURL = mapstorage.NewMap() //инициализируем мапу

	switch {

	case config.FlagDatabaseDSN != "": //работаем с DB если есть переменная среды или флаг ком строки
		// Инициализируем базу данных Postgres
		var DatabaseDSN *pgconn.Config
		DatabaseDSN, err := pgconn.ParseConfig(config.FlagDatabaseDSN)
		if err != nil {
			stdLog.Fatal("Ошибка парсинга строки инициализации БД Postgres")
		}

		storage, err := dbstorage.NewDB(DatabaseDSN)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer storage.Close()
		log.Info("database connection", slog.String("database", DatabaseDSN.Database))

		// инициализируем обработчики
		r.Get("/{alias}", geturl.GetURL(log, storage))
		r.Get("/ping", getpingdb.GetPingDB(log, storage))
		r.Get("/api/user/urls", getuserallurl.GetUserAllURL(log, storage))
		r.Post("/", saveurl.PostURL(log, storage))
		r.Post("/api/shorten", saveurljson.PostURLjson(log, storage))
		r.Post("/api/shorten/batch", batch.PostBatch(log, storage))
		r.Delete("/api/user/urls", deletedflag.DeleteShortURL(log, storage.DelChan))
		trustedSubnet, err := startTrustedSubnet(config.FlagTrustedSubnets)
		if err == nil {
			r.Get("/api/internal/stats", stats.GetStats(log, storage, &trustedSubnet))
		} else {
			log.Info("Доверенная подсеть не задана")
		}

	case config.FlagFileStoragePath != "": //работаем с json файлом если нет DB

		//Читаем в мапу из файла с json записями url
		readFromJSONFile, err := jsonstorage.NewJSONFileReader(config.FlagFileStoragePath)
		if err != nil {
			fmt.Println("Не удалось открыть резервный файл с json сокращениями", config.FlagFileStoragePath)
			return
		}
		log.Info("json file connection", slog.String("file", config.FlagFileStoragePath))

		//mapURL.URLStorage,
		err = readFromJSONFile.ReadToMap(mapURL.URLStorage)
		if err != nil {
			fmt.Println("Не удалось прочитать файл с json сокращениями")
		}
		// инициализируем обработчики
		r.Get("/{alias}", geturl.GetURL(log, mapURL))
		r.Get("/api/user/urls", getuserallurl.GetUserAllURL(log, mapURL))
		r.Post("/", saveurl.PostURL(log, mapURL))
		r.Post("/api/shorten", saveurljson.PostURLjson(log, mapURL))
		r.Post("/api/shorten/batch", batch.PostBatch(log, mapURL))
		r.Delete("/api/user/urls", deletedflag.DeleteShortURL(log, mapURL.DelChan))
		trustedSubnet, err := startTrustedSubnet(config.FlagTrustedSubnets)
		if err == nil {
			r.Get("/api/internal/stats", stats.GetStats(log, mapURL, &trustedSubnet))
		} else {
			log.Info("Доверенная подсеть не задана")
		}

	default: //работаем просто с мапой
		r.Get("/{alias}", geturl.GetURL(log, mapURL))
		r.Get("/api/user/urls", getuserallurl.GetUserAllURL(log, mapURL))
		r.Post("/", saveurl.PostURL(log, mapURL))
		r.Post("/api/shorten", saveurljson.PostURLjson(log, mapURL))
		r.Post("/api/shorten/batch", batch.PostBatch(log, mapURL))
		r.Delete("/api/user/urls", deletedflag.DeleteShortURL(log, mapURL.DelChan))
		trustedSubnet, err := startTrustedSubnet(config.FlagTrustedSubnets)
		if err == nil {
			r.Get("/api/internal/stats", stats.GetStats(log, mapURL, &trustedSubnet))
		} else {
			log.Info("Доверенная подсеть не задана")
		}
	}

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
	if config.FlagServerTLS {
		log.Info("Starting TLS server")
		err := config.GenerateTLS() //генерируем сертификат и ключ
		if err != nil {
			stdLog.Fatal("Ошибка при генерации TLS конфигурации", err.Error())
			return
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
	//Здесь можно добавить запись в файл или базу данных о завершении работы сервера
}

// startTrustedSubnet проверяет наличие подсети в переменной окружения TRUSTED_SUBNET
// если подсеть задана, то возвращает структуру TrustedSubnet
// если подсеть не задана, то возвращает ошибку
func startTrustedSubnet(flagTrustedSubnets string) (models.TrustedSubnet, error) {
	if flagTrustedSubnets != "" {
		trustedSubnet, err := models.NewTrustedSubnet(flagTrustedSubnets)
		if err != nil {
			stdLog.Fatalf("не удалось распарсить переменную окружения TRUSTED_SUBNET: %v", err)
		}
		return trustedSubnet, nil
	}
	return models.TrustedSubnet{}, fmt.Errorf("подсеть не задана")
}
