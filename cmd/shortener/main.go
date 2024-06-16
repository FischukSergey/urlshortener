package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/geturl"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/saveurl"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
	"github.com/go-chi/chi"
)

// var URLStorage = map[string]string{ // временное хранилище URLов
// 	"practicum": "https://practicum.yandex.ru/",
// 	"map":       "https://golangify.com/map",
// }

func main() {
	var log = slog.New( //инициализируем логгер
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	var mapURL = mapstorage.NewMap() //инициализируем мапу

	config.ParseFlags() //инициализируем флаги/переменные окружения конфигурации сервера

	r := chi.NewRouter()  //инициализируем роутер

	r.Get("/{alias}", geturl.GetURL(log, mapURL))
	r.Post("/", saveurl.PostURL(log, mapURL))

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
