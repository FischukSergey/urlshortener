package mapstorage

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/logger"
	"github.com/FischukSergey/urlshortener.git/internal/models"
	"github.com/FischukSergey/urlshortener.git/internal/storage/jsonstorage"
)

// DataStore структура для хранения данных
type DataStore struct {
	URLStorage map[string]config.URLWithUserID //хранилище данных
	DelChan    chan config.DeletedRequest      //канал для записи отложенных запросов на удаление
	mx         sync.RWMutex
}

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

// NewMap() инициализация мапы с мьютексом и каналом fan-in
func NewMap() *DataStore {

	instance := &DataStore{
		URLStorage: make(map[string]config.URLWithUserID, 10000),
		DelChan:    make(chan config.DeletedRequest, 1024), //устанавливаем каналу буфер
	}
	go instance.flushDeletes() //горутина канала fan-in

	return instance
}

// flushMessages постоянно отправляет несколько сообщений в хранилище с определённым интервалом
func (ds *DataStore) flushDeletes() {
	// будем отправлять сообщения, накопленные за последние 10 секунд
	ticker := time.NewTicker(10 * time.Second)

	var delmsges []config.DeletedRequest

	for {
		select {
		case msg := <-ds.DelChan:
			// добавим сообщение в слайс для последующей отправки на удаление
			delmsges = append(delmsges, msg)
		case <-ticker.C:
			// подождём, пока придёт хотя бы одно сообщение
			if len(delmsges) == 0 {
				continue
			}
			//отправим на удаление все пришедшие сообщения одновременно
			err := ds.DeleteBatch(context.TODO(), delmsges...)
			if err != nil {
				log.Debug("cannot delete messages", logger.Err(err))
				// не будем стирать сообщения, попробуем отправить их чуть позже
				continue
			}
			// сотрём успешно отосланные сообщения
			delmsges = nil
		}
	}
}

//реализация методов записи и чтения из мапы с синхронизацией

// GetStorageURL(alias string) метод получения записи из хранилища
// возвращает URL и True для успешного поиска (string, bool)
func (ds *DataStore) GetStorageURL(_ context.Context, alias string) (string, bool) {
	ds.mx.RLock()
	defer ds.mx.RUnlock()
	val, ok := ds.URLStorage[alias]
	if ok && val.DeleteFlag {
		return val.OriginalURL, false //если алиас есть, но помечен на удаление
	}

	return val.OriginalURL, ok
}

// SaveStorageURL(alias, URL string) метод записи в хранилище
func (ds *DataStore) SaveStorageURL(ctx context.Context, saveURL []config.SaveShortURL) error {

	ds.mx.Lock() //блокируем мапу
	defer ds.mx.Unlock()
	for _, s := range saveURL {
		// пишем в мапу
		ds.URLStorage[s.ShortURL] = config.URLWithUserID{
			OriginalURL: s.OriginalURL,
			UserID:      s.UserID,
		}
	}

	if config.FlagFileStoragePath != "" { //открываем файл для записи
		jsonDB, err := jsonstorage.NewJSONFileWriter(config.FlagFileStoragePath)
		if err != nil {
			return fmt.Errorf("%w. Error opening the file: %s ", err, config.FlagFileStoragePath)
		}
		defer func() {
			err := jsonDB.Close()
			if err != nil {
				log.Error("Error close file", logger.Err(err))
			}
		}()
		for _, s := range saveURL {
			//пишем в текстовый файл json строку
			if err = jsonDB.Write(s); err != nil {
				log.Error("Error writing to the file 'short-url-db.json'", logger.Err(err))
			}
		}
	}

	return nil
}

// GetAllUserURL() получение всех записей пользователя
func (ds *DataStore) GetAllUserURL(ctx context.Context, userID int) ([]models.AllURLUserID, error) {
	const op = "mapstorage.GetAllUserURL"
	log = log.With(slog.String("method from", op))
	ds.mx.RLock()
	defer ds.mx.RUnlock()

	var getUserURLs []models.AllURLUserID

	for shortURL, userURL := range ds.URLStorage {
		if userURL.UserID == userID && !userURL.DeleteFlag {
			getUserURLs = append(getUserURLs, models.AllURLUserID{
				ShortURL:    shortURL,
				OriginalURL: userURL.OriginalURL,
			})
		}
	}

	return getUserURLs, nil
}

// DeleteBatch метод удаления записей по списку сокращенных URl сделанных определенным пользователем
func (ds *DataStore) DeleteBatch(ctx context.Context, delmsges ...config.DeletedRequest) error {
	ds.mx.Lock() //блокируем мапу
	defer ds.mx.Unlock()
	count := 0 //счетчик удачных удалений
	for _, delmsg := range delmsges {
		val, ok := ds.URLStorage[delmsg.ShortURL]
		if ok && val.UserID == delmsg.UserID { //переписываем флаг на признак 'удален'
			ds.URLStorage[delmsg.ShortURL] = config.URLWithUserID{
				OriginalURL: val.OriginalURL,
				UserID:      val.UserID,
				DeleteFlag:  true,
			}
			count++
		}
	}

	if count > 0 && config.FlagFileStoragePath != "" { //открываем файл для перезаписи
		jsonFile, err := jsonstorage.NewJSONFileReWriter(config.FlagFileStoragePath)
		if err != nil {
			return fmt.Errorf("%w. Error opening the file: %s ", err, config.FlagFileStoragePath)
		}
		defer func() {
			err := jsonFile.Close()
			if err != nil {
				log.Error("Error close file", logger.Err(err))
			}
		}()
		if err = jsonFile.DeleteFlag(ds.URLStorage); err != nil {
			log.Error("Error delete flag", logger.Err(err))
		}
	}

	return nil
}

// GetStats() метод получения статистики по количеству пользователей и сокращенных URL
func (ds *DataStore) GetStats(ctx context.Context) (config.Stats, error) {
	ds.mx.RLock()
	defer ds.mx.RUnlock()

	userIDs := make(map[int]struct{})
	k := 0
	for _, urlWithUserID := range ds.URLStorage {
		if !urlWithUserID.DeleteFlag {
			userIDs[urlWithUserID.UserID] = struct{}{}
			k++
		}
	}
	stats := config.Stats{
		URLs:  k,
		Users: len(userIDs),
	}
	if stats.URLs == 0 && stats.Users == 0 {
		return config.Stats{}, fmt.Errorf("no data in storage")
	}
	return stats, nil
}

// GetPingDB проверяет соединение с базой данных
func (ds *DataStore) GetPingDB() error {
	return nil
}

func (ds *DataStore) Close() {
	close(ds.DelChan)
}

func (ds *DataStore) Ping() error {
	return nil
}
