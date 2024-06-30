package config

import (
	"flag"
	"os"
)

var ipAddr string = "localhost" //адрес сервера
var FlagServerPort string       //адрес сервера и порта
var FlagBaseURL string          //базовый адрес для редиректа
var FlagFileStoragePath string  //базовый путь хранения файла db json

func ParseFlags() {

	defaultRunAddr := ipAddr + ":8080"
	defaultBaseURL := "http://" + defaultRunAddr
	defaultFileStoragePath := "./tmp/short-url-db.json"

	flag.StringVar(&FlagServerPort, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&FlagBaseURL, "b", defaultBaseURL, "base redirect path")
	flag.StringVar(&FlagFileStoragePath, "f", defaultFileStoragePath, "path file json storage")

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
	}
}
