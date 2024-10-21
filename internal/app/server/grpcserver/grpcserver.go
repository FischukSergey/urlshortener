package grpcserver

import (
	"context"
	stdLog "log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"

	"github.com/FischukSergey/urlshortener.git/internal/grpc/handlers"
	"github.com/FischukSergey/urlshortener.git/internal/grpc/interceptors/mwdecrypt"
	"github.com/FischukSergey/urlshortener.git/internal/grpc/interceptors/mwtrustsubnet"
	"github.com/FischukSergey/urlshortener.git/internal/grpc/services"
	"github.com/FischukSergey/urlshortener.git/internal/logger"
	"github.com/FischukSergey/urlshortener.git/internal/storage"
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
	storage, err := storage.InitStorage(log)
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
	//опции для обработки паники в grpc сервере
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			log.Error("gRPC server panic", logger.Err(err))
			return status.Errorf(codes.Internal, "panic: %v", p)
		}),
	}

	// инициализация grpc сервера
	grpcApp := &GRPCServer{
		port: port,
		log:  log,
	}
	grpcApp.gRPCServer = grpc.NewServer(grpc.ChainUnaryInterceptor(
		mwdecrypt.UnaryDecryptInterceptor,                                      //мидлвар для расшифровки токена
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...), //мидлвар для логирования
		mwtrustsubnet.UnaryTrustSubnetInterceptor,                              //мидлвар для проверки доверенной подсети
		recovery.UnaryServerInterceptor(recoveryOpts...),                       //мидлвар для обработки паники
	))

	handlers.Register(grpcApp.gRPCServer, grpcService) //регистрируем хендлеры в grpc сервере

	return &App{
		GRPCServer: grpcApp,
	}
}

// Run запуск grpc сервера
func (app *GRPCServer) MustRun() {
	go func() {
		if err := app.Run(); err != nil {
			app.log.Error("Error starting gRPC server", logger.Err(err))
			panic(err)
		}
	}()
	//graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	app.log.Info("Stopping gRPC server")
	app.gRPCServer.GracefulStop()
	app.log.Info("gRPC server stopped")
}

// Run запуск grpc сервера
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

// InterceptorLogger обертка интерцептора для логирования
// меняем logging.LevelInfo на slog.LevelInfo
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.LevelInfo, msg, fields...)
	})
}
