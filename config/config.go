package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/FischukSergey/urlshortener.git/internal/models"
)

// AliasLength - длина сокращенного URL
const (
	AliasLength int = 8
)

// переменные для парсинга флагов
var (
	IPAddr              string               = "localhost" //адрес сервера
	IPPort              string               = ":8080"     //порт сервера
	FlagServerPort      string                             //адрес сервера и порта
	FlagBaseURL         string                             //базовый адрес для редиректа
	FlagFileStoragePath string                             //базовый путь хранения файла db json
	FlagDatabaseDSN     string                             //наименование базы данных
	FlagServerTLS       bool                               //флаг для запуска сервера с TLS
	FlagFileConfig      string                             //путь к файлу конфигурации JSON
	FlagTrustedSubnets  string                             //подсети, которые могут использовать API
	TrustedSubnet       models.TrustedSubnet               //доверенная подсеть
	FlagGRPC            bool                               //флаг для запуска сервера с GRPC
)

// Config - структура для конфигурации
type Config struct {
	ServerAddress   string `json:"server_address"`    //адрес сервера и порта
	BaseURL         string `json:"base_url"`          //базовый адрес для редиректа
	FileStoragePath string `json:"file_storage_path"` //базовый путь хранения файла db json
	DatabaseDSN     string `json:"database_dsn"`      //наименование базы данных
	TrustedSubnets  string `json:"trusted_subnets"`   //подсети, которые могут использовать API
	ServerTLS       bool   `json:"enable_https"`      //флаг для запуска сервера с TLS
	GRPC            bool   `json:"grpc"`              //флаг для запуска сервера с GRPC
}

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

// Stats структура для хранения статистики
type Stats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// ParseFlags - функция для парсинга флагов
func ParseFlags() {

	defaultRunAddr := IPAddr + IPPort                   //адрес сервера и порта
	defaultBaseURL := "http://" + defaultRunAddr        //базовый адрес для редиректа
	defaultFileStoragePath := "./tmp/short-url-db.json" //базовый путь хранения файла db json
	defaultDatabaseDSN := ""                            //"user=postgres password=postgres host=localhost port=5432 dbname=urlshortdb sslmode=disable"

	flag.StringVar(&FlagServerPort, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&FlagBaseURL, "b", defaultBaseURL, "base redirect path")
	flag.StringVar(&FlagFileStoragePath, "f", defaultFileStoragePath, "path file json storage")
	flag.StringVar(&FlagDatabaseDSN, "d", defaultDatabaseDSN, "name database Postgres")
	flag.BoolVar(&FlagServerTLS, "s", false, "run server with TLS")
	flag.StringVar(&FlagFileConfig, "c", "", "path to config file")
	flag.StringVar(&FlagTrustedSubnets, "t", "", "trusted subnets")
	flag.BoolVar(&FlagGRPC, "g", false, "run server with GRPC")
	flag.Parse()

	//базовые значения конфигурации
	config := Config{
		ServerAddress:   "",
		BaseURL:         "",
		FileStoragePath: "",
		DatabaseDSN:     "",
		ServerTLS:       false,
		TrustedSubnets:  "",
		GRPC:            false,
	}

	//если есть переменная окружения CONFIG, то используем её
	if envFileConfig, ok := os.LookupEnv("CONFIG"); ok {
		FlagFileConfig = envFileConfig
	}
	if FlagFileConfig != "" {
		file, err := os.Open(FlagFileConfig)
		if err != nil {
			log.Fatalf("не удалось открыть файл конфигурации: %v", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("не удалось закрыть файл конфигурации: %v", err)
			}
		}()
		//парсим файл конфигурации
		err = json.NewDecoder(file).Decode(&config)
		if err != nil {
			log.Fatalf("не удалось распарсить файл конфигурации: %v", err)
		}
	}

	//проверяем остальные переменные окружения
	//приоритет переменных окружения выше флагов, флагов выше конфигурации
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		FlagServerPort = envRunAddr
	} else {
		if FlagServerPort == "" {
			FlagServerPort = config.ServerAddress
		}
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		FlagBaseURL = envBaseURL
	} else {
		if FlagBaseURL == "" {
			FlagBaseURL = config.BaseURL
		}
	}

	if envTrustedSubnets := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnets != "" {
		FlagTrustedSubnets = envTrustedSubnets
	} else {
		if FlagTrustedSubnets == "" {
			FlagTrustedSubnets = config.TrustedSubnets
		}
	}

	if envFlagFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		FlagFileStoragePath = envFlagFileStoragePath
	} else {
		if FlagFileStoragePath == "" {
			FlagFileStoragePath = config.FileStoragePath
		}
	}

	envDatabaseDSN, ok := os.LookupEnv("DATABASE_DSN")
	if ok && envDatabaseDSN != "" {
		FlagDatabaseDSN = envDatabaseDSN
	} else {
		if FlagDatabaseDSN == "" {
			FlagDatabaseDSN = config.DatabaseDSN
		}
	}

	if envEnableTLS, ok := os.LookupEnv("ENABLE_HTTPS"); ok && envEnableTLS != "" {
		envEnableTLSBool, err := strconv.ParseBool(envEnableTLS)
		if err != nil {
			log.Fatalf("не удалось распарсить переменную окружения ENABLE_HTTPS: %v", err)
		}
		FlagServerTLS = envEnableTLSBool
	} else {
		if !FlagServerTLS {
			FlagServerTLS = config.ServerTLS
		}
	}

	if envGRPC, ok := os.LookupEnv("ENABLE_GRPC"); ok && envGRPC != "" {
		envGRPCBool, err := strconv.ParseBool(envGRPC)
		if err != nil {
			log.Fatalf("не удалось распарсить переменную окружения ENABLE_GRPC: %v", err)
		}
		FlagGRPC = envGRPCBool
	} else {
		if !FlagGRPC {
			FlagGRPC = config.GRPC
		}
	}
}
