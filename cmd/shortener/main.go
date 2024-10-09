package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/server/grpcserver"
	"github.com/FischukSergey/urlshortener.git/internal/app/server/httpserver"
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

	if config.FlagGRPC {
		application := grpcserver.New(log, config.IPPort)
		application.GRPCServer.MustRun()
	} else {
		httpserver.StartHTTPServer(log)
	}
}
