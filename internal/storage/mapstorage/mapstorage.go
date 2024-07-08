package mapstorage

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/storage/jsonstorage"
)

type DataStore struct {
	mx         sync.RWMutex
	URLStorage map[string]string
}

// NewMap() инициализация мапы с двумя примерами хранения URL для тестов
func NewMap() *DataStore {
	return &DataStore{
		URLStorage: map[string]string{
			// "practicum": "https://practicum.yandex.ru/",
			// "map":       "https://golangify.com/map",
		},
	}
}

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

//реализация методов записи и чтения из мапы с синхронизацией

// GetStorageURL(alias string) метод получения записи из хранилища
// возвращает URL и True для успешного поиска (string, bool)
func (ds *DataStore) GetStorageURL(ctx context.Context, alias string) (string, bool) {
	ds.mx.RLock()
	defer ds.mx.RUnlock()
	val, ok := ds.URLStorage[alias]
	return val, ok
}

// SaveStorageURL(alias, URL string) метод записи в хранилище
func (ds *DataStore) SaveStorageURL(ctx context.Context, saveURL []config.SaveShortURL) error {

	ds.mx.Lock() //блокируем мапу
	defer ds.mx.Unlock()
	for _, s := range saveURL {
		// пишем в мапу
		ds.URLStorage[s.ShortURL] = s.OriginalURL
	}

	if config.FlagFileStoragePath != "" { //открываем файл для записи
		jsonDB, err := jsonstorage.NewJSONFileWriter(config.FlagFileStoragePath)
		if err != nil {
			return fmt.Errorf("%w. Error opening the file: %s ", err, config.FlagFileStoragePath)
		}
		defer jsonDB.Close()
		for _, s := range saveURL {
			//пишем в текстовый файл json строку
			if err = jsonDB.Write(s.ShortURL, s.OriginalURL); err != nil {
				log.Error("Error writing to the file 'short-url-db.json'", err)
			}
		}
	}

	return nil
}

// GetAll() 	может пригодиться
func (ds *DataStore) GetAll() map[string]string {
	ds.mx.RLock()
	defer ds.mx.RUnlock()

	mapCopy := make(map[string]string, len(ds.URLStorage))
	for key, val := range ds.URLStorage {
		mapCopy[key] = val
	}

	return mapCopy
}
