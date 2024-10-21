package mwtrustsubnet

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/trustsubnet"
	"github.com/FischukSergey/urlshortener.git/internal/logger"
)

var log = slog.New( //инициализируем логгер
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

// UnaryTrustSubnetInterceptor проверяет, что запрос пришел из доверенной подсети
func UnaryTrustSubnetInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Info("request from trusted subnet", "subnet", config.FlagTrustedSubnets)
	if info.FullMethod == "/urlshortener.URLShortener/GetStats" { //проверяем, что запрос идет на получение статистики
		//получаем маску IP из метаданных
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Error("GetStats", logger.Err(errors.New("metadata is not provided")))
			return nil, status.Errorf(codes.InvalidArgument, "metadata is not provided")
		}
		mask := md.Get("X-Real-IP")
		if len(mask) == 0 {
			log.Error("GetStats", logger.Err(errors.New("mask is not provided")))
			return nil, status.Errorf(codes.InvalidArgument, "mask is not provided")
		}
		//проверяем, что сеть задана
		if config.FlagTrustedSubnets == "" {
			log.Error("GetStats", logger.Err(errors.New("trusted subnet is not provided")))
			return nil, status.Errorf(codes.InvalidArgument, "trusted subnet is not provided")
		}
		//парсим маску IP
		userIP := net.ParseIP(mask[0])
		if userIP == nil {
			log.Error("GetStats", logger.Err(errors.New("invalid parsed IP address")))
			return nil, status.Errorf(codes.InvalidArgument, "invalid parsed IP address")
		}
		//проверка на доступность IP из доверенной подсети
		trustedSubnet, err := trustsubnet.StartTrustedSubnet(log, config.FlagTrustedSubnets)
		if err != nil {
			log.Error("GetStats", logger.Err(err))
			return nil, status.Errorf(codes.InvalidArgument, "invalid IP address")
		}
		if !trustedSubnet.IsTrusted(userIP) {
			log.Error("GetStats", logger.Err(errors.New("user is not in trusted subnet")))
			return nil, status.Errorf(codes.InvalidArgument, "user is not in trusted subnet")
		}
		//если IP из доверенной подсети, то продолжаем выполнение запроса
		return handler(ctx, req)
	}
	log.Info("trusted subnet interceptor not used")
	return handler(ctx, req)
}
