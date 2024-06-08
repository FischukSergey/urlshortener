package config

import (
	"flag"
	"os"
)

var ipAddr string = "localhost" //адрес сервера
var FlagServerPort string       //адрес сервера и порта
var FlagBaseURL string          //базовый адрес для редиректа

func ParseFlags() {

	defaultRunAddr := ipAddr + ":8080"
	defaultBaseURL := "http://" + defaultRunAddr

	flag.StringVar(&FlagServerPort, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&FlagBaseURL, "b", defaultBaseURL, "base redirect path")

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		FlagServerPort = envRunAddr
	}

	if envBaseUrl := os.Getenv("BASE_URL"); envBaseUrl != "" {
		FlagBaseURL = envBaseUrl
	}
}
