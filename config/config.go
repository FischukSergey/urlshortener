package config

import (
	"flag"
)

var IPAddr string = "localhost" //адрес сервера
var FlagServerPort string       //адрес порта
var FlagBaseURL string          //базовый адрес для редиректа

func ParseFlags() {
	
	defaultRunAddr := IPAddr + ":8080"
	defaultBaseURL := "http://" + defaultRunAddr

	flag.StringVar(&FlagServerPort, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&FlagBaseURL, "b", defaultBaseURL, "base redirect path")
	
	flag.Parse()
}
