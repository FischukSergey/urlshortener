package handlers

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/FischukSergey/urlshortener.git/internal/proto"
)

// serverAPI структура для работы с grpc
type serverAPI struct {
	pb.UnimplementedURLShortenerServer
	shortener URLShortener
}

// URLShortener интерфейс для работы с api
type URLShortener interface {
	GetURL(ctx context.Context, shortURL string) (string, error)
}

// Register регистрирует serverAPI в grpc сервере
func Register(server *grpc.Server, shortener URLShortener) {
	pb.RegisterURLShortenerServer(server, &serverAPI{shortener: shortener})
}

// GetURL получает оригинальный url по короткому
func (s *serverAPI) GetURL(ctx context.Context, req *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	if req.ShortUrl == "" {
		return nil, status.Errorf(codes.InvalidArgument, "shortURL is empty")
	}

	originalURL, err := s.shortener.GetURL(ctx, req.ShortUrl)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "shortURL not found")
	}

	return &pb.GetURLResponse{OriginalUrl: originalURL}, nil
}

/*
// InitStorage инициализирует хранилище
func InitStorage() (*dbstorage.Storage, error) {
	var DatabaseDSN *pgconn.Config
	DatabaseDSN, err := pgconn.ParseConfig(config.FlagDatabaseDSN)
	if err != nil {
		fmt.Errorf("Ошибка парсинга строки инициализации БД Postgres. %v", err)
		return nil, err
	}

	storage, err := dbstorage.NewDB(DatabaseDSN)
	if err != nil {
		fmt.Errorf("Ошибка инициализации БД Postgres. %v", err)
		return nil, err
	}
	fmt.Println("database connection", slog.String("database", DatabaseDSN.Database))
	return storage, nil
}
*/
