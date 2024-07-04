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

	var mapURL = mapstorage.NewMap() //инициализируем мапу

	config.ParseFlags() //инициализируем флаги/переменные окружения конфигурации сервера

	//Читаем в мапу из файла с json записями url
	if config.FlagFileStoragePath != "" {
		readFromJSONFile, err := jsonstorage.NewJSONFileReader(config.FlagFileStoragePath)
		if err != nil {
			fmt.Println("Не удалось открыть резервный файл с json сокращениями", config.FlagFileStoragePath)
			return
		}

		//mapURL.URLStorage,
		err = readFromJSONFile.ReadToMap(mapURL.URLStorage)
		if err != nil {
			fmt.Println("Не удалось прочитать файл с json сокращениями")
		}
	}

	// Инициализируем базу данных Postgres
	var dbConfig = config.DBConfig{
		User:     "postgres",
		Password: "postgres", //заменить на переменную окружения
		Host:     "localhost",
		Port:     "5432",
		Database: config.FlagDatabaseDSN,
	}
	storage, err := dbstorage.NewDB(dbConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer storage.Close()

	fmt.Println("установлено соединение с базой данных", dbConfig.Database)

	r := chi.NewRouter()             //инициализируем роутер
	r.Use(mwlogger.NewMwLogger(log)) //маршрут в middleware за логированием
	r.Use(gzipper.NewMwGzipper(log)) //работа со сжатыми запросами/сжатие ответов
	r.Use(middleware.RequestID)

	r.Get("/{alias}", geturl.GetURL(log, mapURL))
	r.Post("/", saveurl.PostURL(log, mapURL))
	r.Post("/api/shorten", saveurljson.PostURLjson(log, mapURL))
	r.Get("/ping", getpingdb.GetPingDB(log, storage))

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
