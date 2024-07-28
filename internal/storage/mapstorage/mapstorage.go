package mapstorage

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/storage/jsonstorage"
)

type DataStore struct {
	mx         sync.RWMutex
	URLStorage map[string]config.URLWithUserID
	DelChan    chan config.DeletedRequest //канал для записи отложенных запросов на удаление
}

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

// NewMap() инициализация мапы с мьютексом и каналом fan-in
func NewMap() *DataStore {

	instance := &DataStore{
		URLStorage: map[string]config.URLWithUserID{},
		DelChan:    make(chan config.DeletedRequest, 1024), //устанавливаем каналу буфер
	}
	go instance.flushDeletes() //горутина канала fan-in

	return instance
}

// flushMessages постоянно отправляет несколько сообщений в хранилище с определённым интервалом
func (s *DataStore) flushDeletes() {
	// будем отправлять сообщения, накопленные за последние 10 секунд
	ticker := time.NewTicker(10 * time.Second)

	var delmsges []config.DeletedRequest

	for {
		select {
		case msg := <-s.DelChan:
			// добавим сообщение в слайс для последующей отправки на удаление
			delmsges = append(delmsges, msg)
		case <-ticker.C:
			// подождём, пока придёт хотя бы одно сообщение
			if len(delmsges) == 0 {
				continue
			}
			//отправим на удаление все пришедшие сообщения одновременно
			err := s.DeleteBatch(context.TODO(), delmsges...)
			if err != nil {
				log.Debug("cannot save messages", err)
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
	if ok {
		if val.DeleteFlag {
			return val.OriginalURL, false //если алиас есть, но помечен на удаление
		}
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
		defer jsonDB.Close()
		for _, s := range saveURL {
			//пишем в текстовый файл json строку
			if err = jsonDB.Write(s); err != nil {
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
		mapCopy[key] = val.OriginalURL
	}

	return mapCopy
}

// DeleteBatch метод удаления записей по списку сокращенных URl сделанных определенным пользователем
func (ds *DataStore) DeleteBatch(ctx context.Context, delmsges ...config.DeletedRequest) error {
	ds.mx.Lock() //блокируем мапу
	defer ds.mx.Unlock()

	for _, delmsg := range delmsges {
		val, ok := ds.URLStorage[delmsg.ShortURL]
		if ok || val.UserID == delmsg.UserID { //переписываем флаг на признак 'удален'
			ds.URLStorage[delmsg.ShortURL] = config.URLWithUserID{
				DeleteFlag: true,
			}
		}
	}

	return nil
}
