package config

import (
	"flag"
	"os"
)

// AliasLength - длина сокращенного URL
const (
	AliasLength int = 8
)

// переменные для парсинга флагов
var (
	ipAddr              string = "localhost" //адрес сервера
	FlagServerPort      string               //адрес сервера и порта
	FlagBaseURL         string               //базовый адрес для редиректа
	FlagFileStoragePath string               //базовый путь хранения файла db json
	FlagDatabaseDSN     string               //наименование базы данных
	FlagServerTLS       bool                 //флаг для запуска сервера с TLS
)

// DBConfig - структура для конфигурации подключения к БД
type DBConfig struct {
	User     string // = "postgres"
	Password string // = "postgres"
	Host     string // = "localhost"
	Port     string // = "5432"
	Database string // = "urlshortdb"
}

// SaveShortURL - структура для записи сокращенных urlов в БД
type SaveShortURL struct {
	ShortURL    string //сокращенный URL
	OriginalURL string //оригинальный URL
	UserID      int    //идентификатор пользователя
}

// URLWithUserID - структура для записи в мапу
type URLWithUserID struct {
	OriginalURL string //оригинальный URL
	UserID      int    //идентификатор пользователя
	DeleteFlag  bool   //флаг удаления
}

// DeletedRequest - структура для пакетного удаления записей
type DeletedRequest struct {
	ShortURL string //сокращенный URL
	UserID   int    //идентификатор пользователя
}

// ParseFlags - функция для парсинга флагов
func ParseFlags() {

	defaultRunAddr := ipAddr + ":8080"                  //адрес сервера и порта
	defaultBaseURL := "http://" + defaultRunAddr        //базовый адрес для редиректа
	defaultFileStoragePath := "./tmp/short-url-db.json" //базовый путь хранения файла db json
	defaultDatabaseDSN := ""                            //"user=postgres password=postgres host=localhost port=5432 dbname=urlshortdb sslmode=disable"

	flag.StringVar(&FlagServerPort, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&FlagBaseURL, "b", defaultBaseURL, "base redirect path")
	flag.StringVar(&FlagFileStoragePath, "f", defaultFileStoragePath, "path file json storage")
	flag.StringVar(&FlagDatabaseDSN, "d", defaultDatabaseDSN, "name database Postgres")
	flag.BoolVar(&FlagServerTLS, "s", false, "run server with TLS")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		FlagServerPort = envRunAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		FlagBaseURL = envBaseURL
	}

	if envFlagFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		// if envFlagFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFlagFileStoragePath != "" {
		FlagFileStoragePath = envFlagFileStoragePath
	} else {
		FlagFileStoragePath = ""
	}

	envDatabaseDSN, ok := os.LookupEnv("DATABASE_DSN")
	if ok && envDatabaseDSN != "" {
		//if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		FlagDatabaseDSN = envDatabaseDSN
	}

	if envEnableTLS, ok := os.LookupEnv("ENABLE_HTTPS"); ok && envEnableTLS != "" {
		FlagServerTLS = envEnableTLS == "true"
	}
}
