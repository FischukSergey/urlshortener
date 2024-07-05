package config

import (
	"flag"
	"os"
)

var ipAddr string = "localhost" //адрес сервера
var FlagServerPort string       //адрес сервера и порта
var FlagBaseURL string          //базовый адрес для редиректа
var FlagFileStoragePath string  //базовый путь хранения файла db json
var FlagDatabaseDSN string      //наименование базы данных

type DBConfig struct {
	User     string // = "postgres"
	Password string // = "postgres"
	Host     string // = "localhost"
	Port     string // = "5432"
	Database string // = "urlshortdb"
}

func ParseFlags() {

	defaultRunAddr := ipAddr + ":8080"
	defaultBaseURL := "http://" + defaultRunAddr
	defaultFileStoragePath := "./tmp/short-url-db.json"
	defaultDatabaseDSN := "user=postgres password=postgres host=localhost port=5432 dbname=urlshortdb sslmode=disable"

	flag.StringVar(&FlagServerPort, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&FlagBaseURL, "b", defaultBaseURL, "base redirect path")
	flag.StringVar(&FlagFileStoragePath, "f", defaultFileStoragePath, "path file json storage")
	flag.StringVar(&FlagDatabaseDSN, "d", defaultDatabaseDSN, "name database Postgres")

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		FlagServerPort = envRunAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		FlagBaseURL = envBaseURL
	}

	//if envFlagFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
	if envFlagFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFlagFileStoragePath != "" {
		FlagFileStoragePath = envFlagFileStoragePath
	}

	//if envDatabaseDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		FlagDatabaseDSN = envDatabaseDSN
	}
}
