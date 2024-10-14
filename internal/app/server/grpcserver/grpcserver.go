package grpcserver

import (
	"context"
	stdLog "log"
	"log/slog"
	"net"

	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/grpc/handlers"
	"github.com/FischukSergey/urlshortener.git/internal/grpc/interceptors/mwdecrypt"
	"github.com/FischukSergey/urlshortener.git/internal/grpc/services"
	"github.com/FischukSergey/urlshortener.git/internal/logger"
	"github.com/FischukSergey/urlshortener.git/internal/storage/dbstorage"
	"github.com/FischukSergey/urlshortener.git/internal/storage/jsonstorage"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

// GRPCServer структура для работы с grpc
type GRPCServer struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       string
}

// App структура для работы с grpc
type App struct {
	GRPCServer *GRPCServer
}

// New создание нового grpc сервера	с инициализацией хранилища и регистрацией хендлеров
func New(log *slog.Logger, port string) *App {
	// инициализация хранилища
	storage, err := InitStorage(log)
	if err != nil {
		stdLog.Fatal("Error initializing storage", logger.Err(err))
	}
	_, ok := storage.(services.Shortener) //проверка на реализацию интерфейса Shortener
	if !ok {
		log.Error("Storage does not implement the Shortener interface")
		return nil
	}
	grpcService := services.NewGRPCService(log, storage.(services.Shortener)) //создание сервиса для работы с grpc

	//TODO: добавить обработку паники в grpc сервере ()
	
	//опции для логирования в middleware
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent),
	}

	// инициализация grpc сервера
	grpcApp := &GRPCServer{
		port: port,
		log:  log,
	}
	grpcApp.gRPCServer = grpc.NewServer(grpc.ChainUnaryInterceptor(
		mwdecrypt.UnaryDecryptInterceptor, //мидлвар для расшифровки токена
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...), //мидлвар для логирования
	)) //TODO: добавить interceptor
	
	handlers.Register(grpcApp.gRPCServer, grpcService) //регистрируем хендлеры в grpc сервере

	return &App{
		GRPCServer: grpcApp,
	}
}

// Run запуск grpc сервера
func (app *GRPCServer) MustRun() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
func (app *GRPCServer) Run() error {
	lis, err := net.Listen("tcp", app.port)
	if err != nil {
		return err
	}
	app.log.Info("Starting gRPC server on port", slog.String("port", app.port))

	//запускаем обработчик gRPC сообщений
	if err := app.gRPCServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

// InitStorage инициализация хранилища
func InitStorage(log *slog.Logger) (storage interface{}, err error) {
	switch {

	case config.FlagDatabaseDSN != "":
		// работаем с базой данных если есть переменная среды или флаг ком строки
		var DatabaseDSN *pgconn.Config
		DatabaseDSN, err := pgconn.ParseConfig(config.FlagDatabaseDSN)
		if err != nil {
			log.Error("Error parsing database DSN", "error", err)
			return nil, err
		}
		storage, err = dbstorage.NewDB(DatabaseDSN)
		if err != nil {
			log.Error("Error initializing database", "error", err)
			return nil, err
		}
		log.Info("Using database storage", "database", DatabaseDSN.Database)

	case config.FlagFileStoragePath != "": //работаем с json файлом если нет DB
		mapStorage := mapstorage.NewMap()
		//Читаем в мапу из файла с json записями url
		readFromJSONFile, err := jsonstorage.NewJSONFileReader(config.FlagFileStoragePath)
		if err != nil {
			stdLog.Fatal("Не удалось открыть резервный файл с json сокращениями", config.FlagFileStoragePath)
		}
		log.Info("json file connection", slog.String("file", config.FlagFileStoragePath))

		//mapURL.URLStorage,
		err = readFromJSONFile.ReadToMap(mapStorage.URLStorage)
		if err != nil {
			log.Error("Не удалось прочитать файл с json сокращениями", logger.Err(err))
		}
		storage = mapStorage
		log.Info("Using json file storage", slog.String("file", config.FlagFileStoragePath))

	default:
		storage = mapstorage.NewMap()
		log.Info("Using map storage")
	}

	log.Info("Storage initialized", slog.Any("storage", storage))
	return storage, nil
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.LevelInfo, msg, fields...)
	})
}
