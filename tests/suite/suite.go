package suite

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/FischukSergey/urlshortener.git/config"
	pb "github.com/FischukSergey/urlshortener.git/internal/proto"
)

// Suite структура для тестирования grpc сервера
type Suite struct {
	*testing.T
	GrpcClient pb.URLShortenerClient
}

// New создание нового теста
func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	//t.Parallel()
	//основной родительский контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	//закрытие контекста при завершении теста
	t.Cleanup(func() {
		t.Helper()
		cancel()
	})
	grpcAddr := net.JoinHostPort(config.IPAddr, strings.TrimPrefix(config.IPPort, ":"))
	//создание соединения с grpc сервером
	cc, err := grpc.NewClient(grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}
	//создание клиента grpc
	grpcClient := pb.NewURLShortenerClient(cc)

	return ctx, &Suite{T: t, GrpcClient: grpcClient}
}
