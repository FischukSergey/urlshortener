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
	FlagServerTLS       string                             //флаг для запуска сервера с TLS
	FlagFileConfig      string                             //путь к файлу конфигурации JSON
	FlagTrustedSubnets  string                             //подсети, которые могут использовать API
	TrustedSubnet       models.TrustedSubnet               //доверенная подсеть
	FlagGRPC            string                             //флаг для запуска сервера с GRPC
	GRPC                bool                               //флаг для запуска сервера с GRPC
	ServerTLS           bool                               //флаг для запуска сервера с TLS
)

// Config - структура для конфигурации
type Config struct {
	ServerAddress   string `json:"server_address"`    //адрес сервера и порта
	BaseURL         string `json:"base_url"`          //базовый адрес для редиректа
	FileStoragePath string `json:"file_storage_path"` //базовый путь хранения файла db json
	DatabaseDSN     string `json:"database_dsn"`      //наименование базы данных
	TrustedSubnets  string `json:"trusted_subnets"`   //подсети, которые могут использовать API
	GRPC            string `json:"grpc"`              //флаг для запуска сервера с GRPC
	ServerTLS       string `json:"enable_https"`      //флаг для запуска сервера с TLS
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
	defaultGRPC := "false"                              //флаг для запуска сервера с GRPC
	defaultServerTLS := "false"                         //флаг для запуска сервера с TLS
	defaultDatabaseDSN := ""                            //"user=postgres password=postgres host=localhost port=5432 dbname=urlshortdb sslmode=disable"

	flag.StringVar(&FlagServerPort, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&FlagBaseURL, "b", defaultBaseURL, "base redirect path")
	flag.StringVar(&FlagFileStoragePath, "f", defaultFileStoragePath, "path file json storage")
	flag.StringVar(&FlagDatabaseDSN, "d", defaultDatabaseDSN, "name database Postgres")
	flag.StringVar(&FlagServerTLS, "s", defaultServerTLS, "run server with TLS")
	flag.StringVar(&FlagFileConfig, "c", "", "path to config file")
	flag.StringVar(&FlagTrustedSubnets, "t", "", "trusted subnets")
	flag.StringVar(&FlagGRPC, "g", defaultGRPC, "run server with GRPC")
	flag.Parse()

	//базовые значения конфигурации
	config := Config{
		ServerAddress:   "",
		BaseURL:         "",
		FileStoragePath: "",
		DatabaseDSN:     "",
		ServerTLS:       "",
		TrustedSubnets:  "",
		GRPC:            "",
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
	//проверяем на доверенную подсеть

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

	envEnableTLS, ok := os.LookupEnv("ENABLE_HTTPS")
	switch {
	case ok && envEnableTLS != "": //если есть переменная окружения и она не пустая
		envEnableTLSBool, err := strconv.ParseBool(envEnableTLS)
		if err != nil {
			log.Fatalf("не удалось распарсить переменную окружения ENABLE_HTTPS: %v", err)
		}
		ServerTLS = envEnableTLSBool
	case FlagServerTLS != "": //если есть флаг и он не пустой
		ServerTLSBool, err := strconv.ParseBool(FlagServerTLS)
		if err != nil {
			log.Fatalf("не удалось распарсить флаг s-: %v", err)
		}
		ServerTLS = ServerTLSBool
	default: //если нет переменных окружения и флагов, то используем значение из файла конфигурации
		jsonServerTLS, err := strconv.ParseBool(config.ServerTLS)
		if err != nil {
			log.Fatalf("не удалось распарсить значение из файла конфигурации enable_https: %v", err)
		}
		ServerTLS = jsonServerTLS
	}

	envGRPC, ok := os.LookupEnv("ENABLE_GRPC")
	switch {
	case ok && envGRPC != "": //если есть переменная окружения и она не пустая
		envGRPCBool, err := strconv.ParseBool(envGRPC)
		if err != nil {
			log.Fatalf("не удалось распарсить переменную окружения ENABLE_GRPC: %v", err)
		}
		GRPC = envGRPCBool
	case FlagGRPC != "": //если есть флаг и он не пустой
		GRPCBool, err := strconv.ParseBool(FlagGRPC)
		if err != nil {
			log.Fatalf("не удалось распарсить флаг g-: %v", err)
		}
		GRPC = GRPCBool
	default: //если нет переменных окружения и флагов, то используем значение из файла конфигурации
		jsonGRPC, err := strconv.ParseBool(config.GRPC)
		if err != nil {
			log.Fatalf("не удалось распарсить значение из файла конфигурации grpc: %v", err)
		}
		GRPC = jsonGRPC
	}
}
