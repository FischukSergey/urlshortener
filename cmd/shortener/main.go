package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/getpingdb"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/geturl"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/saveurl"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/saveurljson"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/gzipper"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/mwlogger"
	"github.com/FischukSergey/urlshortener.git/internal/storage/dbstorage"
	"github.com/FischukSergey/urlshortener.git/internal/storage/jsonstorage"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	var log = slog.New( //инициализируем логгер
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	config.ParseFlags() //инициализируем флаги/переменные окружения конфигурации сервера

	r := chi.NewRouter()             //инициализируем роутер и middleware
	r.Use(mwlogger.NewMwLogger(log)) //маршрут в middleware за логированием
	r.Use(gzipper.NewMwGzipper(log)) //работа со сжатыми запросами/сжатие ответов
	r.Use(middleware.RequestID)

	var mapURL = mapstorage.NewMap() //инициализируем мапу

	switch {

	case config.FlagDatabaseDSN != "": //работаем с DB если есть переменная среды или флаг ком строки
		// Инициализируем базу данных Postgres
		var dbConfig = config.DBConfig{
			User:     "postgres",
			Password: "postgres", //TODO заменить на переменную окружения
			Host:     "127.0.0.1",
			Port:     "5432",
			Database: config.FlagDatabaseDSN,
		}
		storage, err := dbstorage.NewDB(dbConfig)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer storage.Close()
		log.Info("database connection", slog.String("database", dbConfig.Database))

		// инициализируем обработчики
		r.Get("/{alias}", geturl.GetURL(log, storage))
		r.Post("/", saveurl.PostURL(log, storage))
		r.Post("/api/shorten", saveurljson.PostURLjson(log, storage))
		r.Get("/ping", getpingdb.GetPingDB(log, storage))

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
		r.Post("/", saveurl.PostURL(log, mapURL))
		r.Post("/api/shorten", saveurljson.PostURLjson(log, mapURL))

	default: //работаем просто с мапой
		r.Get("/{alias}", geturl.GetURL(log, mapURL))
		r.Post("/", saveurl.PostURL(log, mapURL))
		r.Post("/api/shorten", saveurljson.PostURLjson(log, mapURL))

	}

	//запускаем сервер
	srv := &http.Server{
		Addr:         config.FlagServerPort,
		Handler:      r,
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 4 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Info("Initializing server", slog.String("address", srv.Addr))

	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
	}
}
