package jsonstorage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/FischukSergey/urlshortener.git/config"
)

type JSONRaw struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	DeleteFlag  bool   `json:"delete_flag"`
}

type JSONFileWriter struct {
	file       *os.File
	JSONWriter *json.Encoder
}

// NewJSONFileWriter() создаем объект с открытым файлом для записи
func NewJSONFileWriter(fileName string) (*JSONFileWriter, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &JSONFileWriter{
		file:       file,
		JSONWriter: json.NewEncoder(file),
	}, nil
}

// Write() метод для записи новой json строки в файл
func (fs *JSONFileWriter) Write(s config.SaveShortURL) error {
	raw := JSONRaw{
		ShortURL:    s.ShortURL,
		OriginalURL: s.OriginalURL,
		UUID:        strconv.Itoa(s.UserID),
	}

	return fs.JSONWriter.Encode(raw)
}

func (fs *JSONFileWriter) Close() error {
	return fs.file.Close()
}

// Загрузка файла с json-записями в мапу
type JSONFileReader struct {
	file    *os.File
	ScanRaw *bufio.Scanner
}

// инициализация объекта file и сканера строк к нему
func NewJSONFileReader(filename string) (*JSONFileReader, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &JSONFileReader{
		file:    file,
		ScanRaw: bufio.NewScanner(file),
	}, nil
}

// Метод ReadToMap() принимает мапу(пустую) и заполняет ее данными из json файла
func (fr *JSONFileReader) ReadToMap(mapURL map[string]config.URLWithUserID) error { //(map[string]string, error) { //чтение файла в мапу до запуска сервера, поэтому работаем без mutex
	defer fr.file.Close()

	for fr.ScanRaw.Scan() { //построчно читаем, декодируем и пишем в мапу
		data := fr.ScanRaw.Bytes()
		mapLine := &JSONRaw{}
		err := json.Unmarshal(data, &mapLine)
		if err != nil {
			fmt.Println("Не удалось декодировать строку:", data)
			return err //nil, err
		}
		if _, ok := mapURL[mapLine.ShortURL]; ok {
			fmt.Printf("Алиас %s дублируется:\n", mapLine.ShortURL)
		} else {
			userID, err := strconv.Atoi(mapLine.UUID)
			if err != nil { //если ID  не порядковый номер
				userID = -1
			}
			mapURL[mapLine.ShortURL] = config.URLWithUserID{
				OriginalURL: mapLine.OriginalURL,
				UserID:      userID,
			}
		}
	}
	return nil //mapURL, nil
}

func (fr *JSONFileReader) Close() error {
	return fr.file.Close()
}

// {"uuid":"1","short_url":"4rSPg8ap","original_url":"http://yandex.ru"}
// {"uuid":"2","short_url":"edVPg3ks","original_url":"http://ya.ru"}
// {"uuid":"3","short_url":"dG56Hqxm","original_url":"http://practicum.yandex.ru"}
