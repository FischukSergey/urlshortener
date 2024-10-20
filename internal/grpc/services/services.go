package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/logger"
	"github.com/FischukSergey/urlshortener.git/internal/models"
	"github.com/FischukSergey/urlshortener.git/internal/storage/dbstorage"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
	"github.com/FischukSergey/urlshortener.git/internal/utilitys"
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
	SaveStorageURL(ctx context.Context, saveURL []config.SaveShortURL) error
	GetPingDB() error
	GetAllUserURL(ctx context.Context, userID int) ([]models.AllURLUserID, error)
	GetStats(ctx context.Context) (config.Stats, error)
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

// PostURL создает короткую ссылку
func (s *ShortenerService) PostURL(ctx context.Context, originalURL string, id int) (string, error) {
	s.log.Debug("PostURL", "originalURL", originalURL)
	//генерируем произвольный алиас длины {aliasLength}
	alias := utilitys.NewRandomString(config.AliasLength)
	//создаем слайс с короткой ссылкой
	saveURL := []config.SaveShortURL{
		{
			ShortURL:    alias,
			OriginalURL: originalURL,
			UserID:      id,
		},
	}
	//сохраняем в хранилище
	err := s.shortener.SaveStorageURL(ctx, saveURL)
	if err != nil {
		s.log.Error("PostURL", logger.Err(err))
		return "", err
	}
	fullURL := fmt.Sprintf("%s/%s", config.FlagBaseURL, alias)
	return fullURL, nil
}

// PostBatch создает множественные короткие ссылки
func (s *ShortenerService) PostBatch(ctx context.Context,
	requests []struct {
		CorrelationID string
		OriginalURL   string
	}, id int) ([]struct {
	CorrelationID string
	ShortURL      string
}, error) {

	s.log.Debug("PostBatch", "requests", requests)
	saveURL := make([]config.SaveShortURL, 0, len(requests))
	responses := make([]struct {
		CorrelationID string
		ShortURL      string
	}, 0, len(requests))
	for _, request := range requests {
		alias := utilitys.NewRandomString(config.AliasLength)
		responses = append(responses, struct {
			CorrelationID string
			ShortURL      string
		}{
			CorrelationID: request.CorrelationID,
			ShortURL:      alias,
		})
		saveURL = append(saveURL, config.SaveShortURL{
			ShortURL:    alias,
			OriginalURL: request.OriginalURL,
			UserID:      id,
		})
	}
	err := s.shortener.SaveStorageURL(ctx, saveURL)
	if err != nil {
		s.log.Error("PostBatch", logger.Err(err))
		return nil, err
	}
	return responses, nil
}

// Ping проверяет соединение с базой данных
func (s *ShortenerService) Ping(ctx context.Context) error {
	s.log.Debug("Ping")

	err := s.shortener.GetPingDB()
	if err != nil {
		s.log.Error("GetPingDB", logger.Err(err))
		return err
	}
	return nil
}

// GetUserURLs получает все URL пользователя
func (s *ShortenerService) GetUserURLs(ctx context.Context, id int) ([]struct {
	ShortURL    string
	OriginalURL string
}, error) {
	s.log.Debug("GetUserURLs", "id", id)
	urls, err := s.shortener.GetAllUserURL(ctx, id)
	if err != nil {
		s.log.Error("GetUserURLs", logger.Err(err))
		return nil, err
	}
	responses := make([]struct {
		ShortURL    string
		OriginalURL string
	}, 0, len(urls))
	for _, url := range urls {
		responses = append(responses, struct {
			ShortURL    string
			OriginalURL string
		}{ShortURL: url.ShortURL, OriginalURL: url.OriginalURL})
	}
	return responses, nil
}

// DeleteBatch удаляет URL пользователя
func (s *ShortenerService) DeleteBatch(ctx context.Context, shortUrls []string, id int) error {
	s.log.Debug("DeleteBatch", "shortUrls", shortUrls)
	switch s.shortener.(type) {
	case *dbstorage.Storage:
		for _, url := range shortUrls {
			s.shortener.(*dbstorage.Storage).DelChan <- config.DeletedRequest{ShortURL: url, UserID: id}
		}
	case *mapstorage.DataStore:
		for _, url := range shortUrls {
			s.shortener.(*mapstorage.DataStore).DelChan <- config.DeletedRequest{ShortURL: url, UserID: id}
		}
	}
	return nil
}

// GetStats получает статистику
func (s *ShortenerService) GetStats(ctx context.Context, _ string) (config.Stats, error) {
	s.log.Debug("service GetStats")
	//получаем статистику если IP из доверенной подсети
	stats, err := s.shortener.GetStats(ctx)
	if err != nil {
		s.log.Error("GetStats", logger.Err(err))
		return config.Stats{}, err
	}
	return stats, nil
}
