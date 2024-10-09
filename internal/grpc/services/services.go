package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/FischukSergey/urlshortener.git/internal/logger"
)

var (
	ErrNotFound = errors.New("not found")
)

// ShortenerService структура для работы с api
type ShortenerService struct {
	shortener Shortener
	log       *slog.Logger
}

// Shortener интерфейс для работы с хранилищем
type Shortener interface {
	GetStorageURL(ctx context.Context, alias string) (string, bool)
}

func NewGRPCService(log *slog.Logger, shortener Shortener) *ShortenerService {
	return &ShortenerService{shortener: shortener, log: log}
}

// GetURL получает оригинальный url по короткому из хранилища
func (s *ShortenerService) GetURL(ctx context.Context, shortURL string) (string, error) {
	s.log.Debug("GetURL", "shortURL", shortURL)
	originalURL, ok := s.shortener.GetStorageURL(ctx, shortURL)
	if !ok {
		s.log.Error("GetURL", logger.Err(ErrNotFound))
		return "", ErrNotFound
	}
	return originalURL, nil
}
