package handlers

import (
	"context"
	"errors"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/grpc/interceptors/mwdecrypt"
	pb "github.com/FischukSergey/urlshortener.git/internal/proto"
	"github.com/FischukSergey/urlshortener.git/internal/storage/dbstorage"
)

// serverAPI структура для работы с grpc
type serverAPI struct {
	pb.UnimplementedURLShortenerServer
	shortener URLShortener
}

// URLShortener интерфейс для работы с api
type URLShortener interface {
	GetURL(ctx context.Context, shortURL string) (string, error)
	PostURL(ctx context.Context, originalURL string, id int) (string, error)
	PostBatch(ctx context.Context,
		requests []struct {
			CorrelationID string
			OriginalURL   string
		},
		id int) ([]struct {
		CorrelationID string
		ShortURL      string
	}, error)
	Ping(ctx context.Context) error
	GetUserURLs(ctx context.Context, id int) ([]struct {
		ShortURL    string
		OriginalURL string
	}, error)
	DeleteBatch(ctx context.Context, shortUrls []string, id int) error
	GetStats(ctx context.Context, mask string) (config.Stats, error)
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

// PostURL создает короткую ссылку
func (s *serverAPI) PostURL(ctx context.Context, req *pb.PostURLRequest) (*pb.PostURLResponse, error) {

	if req.OriginalUrl == "" {
		return nil, status.Errorf(codes.InvalidArgument, "originalURL is empty")
	}
	//проверяем валидность url
	if _, err := url.ParseRequestURI(req.OriginalUrl); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request URL")
	}
	//получаем id пользователя из контекста
	id := ctx.Value(mwdecrypt.CtxKeyUserGrpc)
	//создаем короткую ссылку
	fullURL, err := s.shortener.PostURL(ctx, req.OriginalUrl, id.(int))
	if err != nil {
		if errors.Is(err, dbstorage.ErrURLExists) {
			return nil, status.Errorf(codes.AlreadyExists, "URL already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create short URL")
	}
	return &pb.PostURLResponse{BaseUrl: fullURL}, nil
}

// PostBatch создает множественные короткие ссылки
func (s *serverAPI) PostBatch(ctx context.Context, req *pb.PostBatchRequest) (*pb.PostBatchResponse, error) {
	//получаем id пользователя из контекста
	id := ctx.Value(mwdecrypt.CtxKeyUserGrpc)
	//проверяем, что массив не пуст
	if len(req.Requests) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "requests are empty")
	}
	//создаем массив запросов
	var requests []struct {
		CorrelationID string
		OriginalURL   string
	}
	//проверяем валидность url и записываем в массив
	for _, req := range req.Requests {
		if req.OriginalUrl == "" {
			return nil, status.Errorf(codes.InvalidArgument, "originalURL is empty")
		}
		if _, err := url.ParseRequestURI(req.OriginalUrl); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid request URL")
		}
		requests = append(requests, struct {
			CorrelationID string
			OriginalURL   string
		}{
			CorrelationID: req.CorrelationId,
			OriginalURL:   req.OriginalUrl,
		})
	}
	//создаем множественные короткие ссылки
	shortURLs, err := s.shortener.PostBatch(ctx, requests, id.(int))
	if err != nil {
		if errors.Is(err, dbstorage.ErrURLExists) {
			return nil, status.Errorf(codes.AlreadyExists, "URL already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create short URL")
	}
	//создаем массив ответов
	var responses []*pb.Response
	//записываем в массив ответов
	for _, shortURL := range shortURLs {
		responses = append(responses, &pb.Response{CorrelationId: shortURL.CorrelationID, ShortUrl: shortURL.ShortURL})
	}

	return &pb.PostBatchResponse{Responses: responses}, nil
}

// Ping проверяет соединение с базой данных
func (s *serverAPI) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.shortener.Ping(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to ping database")
	}
	return &pb.PingResponse{}, nil
}

// GetUserURLs получает все URL пользователя
func (s *serverAPI) GetUserURLs(ctx context.Context, req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	id := ctx.Value(mwdecrypt.CtxKeyUserGrpc)

	if id == -1 {
		return nil, status.Errorf(codes.Unauthenticated, "token is not provided")
	}
	urls, err := s.shortener.GetUserURLs(ctx, id.(int))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user URLs")
	}
	var pbURLs []*pb.URL
	for _, url := range urls {
		pbURLs = append(pbURLs, &pb.URL{ShortUrl: url.ShortURL, OriginalUrl: url.OriginalURL})
	}
	return &pb.GetUserURLsResponse{Urls: pbURLs}, nil
}

// DeleteUserURLs удаляет множественные URL пользователя
func (s *serverAPI) DeleteUserURLs(ctx context.Context, req *pb.DeleteUserURLsRequest) (*pb.DeleteUserURLsResponse, error) {
	id := ctx.Value(mwdecrypt.CtxKeyUserGrpc)
	if id == -1 {
		return nil, status.Errorf(codes.Unauthenticated, "token is not provided")
	}
	if len(req.ShortUrls) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "shortURLs are empty")
	}

	err := s.shortener.DeleteBatch(ctx, req.ShortUrls, id.(int))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user URLs")
	}
	return &pb.DeleteUserURLsResponse{}, nil
}

// Stats получает статистику
func (s *serverAPI) GetStats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	stats, err := s.shortener.GetStats(ctx, "")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid IP address or subnet")
	}
	return &pb.StatsResponse{Urls: int32(stats.URLs), Users: int32(stats.Users)}, nil
}
