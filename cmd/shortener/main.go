package main

import (
	"fmt"
	stdLog "log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
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
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/gzipper"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/mwlogger"
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

	default: //работаем просто с мапой
		r.Get("/{alias}", geturl.GetURL(log, mapURL))
		r.Get("/api/user/urls", getuserallurl.GetUserAllURL(log, mapURL))
		r.Post("/", saveurl.PostURL(log, mapURL))
		r.Post("/api/shorten", saveurljson.PostURLjson(log, mapURL))
		r.Post("/api/shorten/batch", batch.PostBatch(log, mapURL))
		r.Delete("/api/user/urls", deletedflag.DeleteShortURL(log, mapURL.DelChan))
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

	if config.FlagServerTLS {
		log.Info("Starting TLS server")
		err := config.GenerateTLS()
		if err != nil {
			stdLog.Fatal("Ошибка при генерации TLS конфигурации", err.Error())
			return
		}
		if err := srv.ListenAndServeTLS("server.crt", "server.key"); err != nil {
			stdLog.Fatal("Ошибка при запуске TLS сервера", err.Error())
		}
	} else {
		if err := srv.ListenAndServe(); err != nil {
			stdLog.Fatal("Ошибка при запуске сервера", err.Error())
		}
	}
}
