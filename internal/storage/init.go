package storage

import (
	stdLog "log"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/logger"
	"github.com/FischukSergey/urlshortener.git/internal/storage/dbstorage"
	"github.com/FischukSergey/urlshortener.git/internal/storage/jsonstorage"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
)

// InitStorage инициализация хранилища
func InitStorage(log *slog.Logger) (storage interface{}, err error) {
	switch {

	case config.FlagDatabaseDSN != "":
		// работаем с базой данных если есть переменная среды или флаг ком строки
		var DatabaseDSN *pgconn.Config
		DatabaseDSN, err := pgconn.ParseConfig(config.FlagDatabaseDSN)
		if err != nil {
			log.Error("Error parsing database DSN", "error", err)
			return nil, err
		}
		storage, err = dbstorage.NewDB(DatabaseDSN)
		if err != nil {
			log.Error("Error initializing database", "error", err)
			return nil, err
		}
		log.Info("Using database storage", "database", DatabaseDSN.Database)

	case config.FlagFileStoragePath != "": //работаем с json файлом если нет DB
		mapStorage := mapstorage.NewMap()
		//Читаем в мапу из файла с json записями url
		readFromJSONFile, err := jsonstorage.NewJSONFileReader(config.FlagFileStoragePath)
		if err != nil {
			stdLog.Fatal("Не удалось открыть резервный файл с json сокращениями", config.FlagFileStoragePath)
		}
		log.Info("json file connection", slog.String("file", config.FlagFileStoragePath))

		//mapURL.URLStorage,
		err = readFromJSONFile.ReadToMap(mapStorage.URLStorage)
		if err != nil {
			log.Error("Не удалось прочитать файл с json сокращениями", logger.Err(err))
		}
		storage = mapStorage
		log.Info("Using json file storage", slog.String("file", config.FlagFileStoragePath))

	default:
		storage = mapstorage.NewMap()
		log.Info("Using map storage")
	}

	log.Info("Storage initialized", slog.Any("storage", storage))
	return storage, nil
}
