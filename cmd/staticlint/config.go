package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// константа для пути к файлу конфигурации
const configPath = "staticcheck.yaml"

// структура для хранения данных конфигурации
type ConfigData struct {
	Staticcheck []string `yaml:"checks"` // список проверок, которые нужно выполнять
}

// NewConfig читает файл конфигурации и возвращает структуру ConfigData
func NewConfig() *ConfigData {
	appfile, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("error getting executable path: %w", err))
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), configPath))
	if err != nil {
		panic(fmt.Errorf("error reading config file: %w", err))
	}
	var config ConfigData
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(fmt.Errorf("error unmarshalling config file: %w", err))
	}
	return &config
}
