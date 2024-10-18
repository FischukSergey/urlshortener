package config

import (
	"context"

	"github.com/FischukSergey/urlshortener.git/internal/models"
)

// URLStorage интерфейс методов для работы с хранилищем URL
type URLStorage interface {
	Close()
	DeleteBatch(ctx context.Context, delmsges ...DeletedRequest) error
	GetAllUserURL(ctx context.Context, userID int) ([]models.AllURLUserID, error)
	GetPingDB() error
	GetStats(ctx context.Context) (Stats, error)
	GetStorageURL(ctx context.Context, alias string) (string, bool)
	SaveStorageURL(ctx context.Context, saveURL []SaveShortURL) error
}
