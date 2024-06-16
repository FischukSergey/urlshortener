package mapstorage

import (
	"sync"
)

type DataStore struct {
	mx         sync.RWMutex
	URLStorage map[string]string
}

// NewMap() инициализация мапы с двумя примерами хранения URL для тестов
func NewMap() *DataStore {
	return &DataStore{
		URLStorage: map[string]string{
			"practicum": "https://practicum.yandex.ru/",
			"map":       "https://golangify.com/map",
		},
	}
}

//реализация методов записи и чтения из мапы с синхронизацией

// GetStorageURL(alias string) метод получения записи из хранилища
// возвращает URL и True для успешного поиска (string, bool)
func (ds *DataStore) GetStorageURL(alias string) (string, bool) {
	ds.mx.RLock()
	defer ds.mx.RUnlock()
	val, ok := ds.URLStorage[alias]
	return val, ok
}

//SaveStorageURL(alias, URL string) метод записи в хранилище
func (ds *DataStore) SaveStorageURL(alias, URL string) {
	ds.mx.Lock()
	defer ds.mx.Unlock()
	ds.URLStorage[alias]=URL
}

//GetAll() 	может пригодиться
func (ds *DataStore) GetAll() map[string]string {
ds.mx.RLock()
defer ds.mx.RUnlock()

mapCopy := make(map[string]string, len(ds.URLStorage))
for key, val := range ds.URLStorage {
mapCopy[key] = val
}

return mapCopy
}