package jsonstorage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/FischukSergey/urlshortener.git/config"
)

// JSONRaw структура для записи в json файл
type JSONRaw struct {
	UUID        string `json:"uuid"`         //ID пользователя
	ShortURL    string `json:"short_url"`    //сокращенный URL
	OriginalURL string `json:"original_url"` //оригинальный URL
	DeleteFlag  bool   `json:"delete_flag"`  //флаг на удаление
}

// JSONFileWriter структура для записи в json файл
type JSONFileWriter struct {
	file       *os.File
	JSONWriter *json.Encoder
}

// JSONFileReWriter структура для записи в json файл
type JSONFileReWriter struct {
	file       *os.File
	JSONWriter *json.Encoder
	//JSONReader *json.Decoder
	//ScanRaw    *bufio.Scanner
}

// NewJSONFileWriter создаем объект с открытым файлом для записи
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

// NewJSONFileReWriter() создаем объект с открытым пустым файлом
func NewJSONFileReWriter(fileName string) (*JSONFileReWriter, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &JSONFileReWriter{
		file:       file,
		JSONWriter: json.NewEncoder(file),
		//JSONReader: json.NewDecoder(file),
		//ScanRaw:    bufio.NewScanner(file),
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

// Close закрывает файл
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
			continue
		}
		userID, err := strconv.Atoi(mapLine.UUID)
		if err != nil { //если ID  не порядковый номер
			userID = -1
		}
		mapURL[mapLine.ShortURL] = config.URLWithUserID{
			OriginalURL: mapLine.OriginalURL,
			UserID:      userID,
			DeleteFlag:  mapLine.DeleteFlag,
		}

	}
	return nil //mapURL, nil
}

// Close закрывает файл
func (fr *JSONFileReader) Close() error {
	return fr.file.Close()
}

// DeleteFlag метод помечает на удаление все запрошенные пользователем записи
func (rr *JSONFileReWriter) DeleteFlag(mapLines map[string]config.URLWithUserID) error {

	for mapLine := range mapLines {
		raw := JSONRaw{
			ShortURL:    mapLine,
			OriginalURL: mapLines[mapLine].OriginalURL,
			UUID:        strconv.Itoa(mapLines[mapLine].UserID),
			DeleteFlag:  mapLines[mapLine].DeleteFlag,
		}
		err := rr.JSONWriter.Encode(raw)
		if err != nil {
			return fmt.Errorf("no write json row with ShortURL: %s", mapLine)
		}
	}
	return nil
}

// Close закрывает файл
func (rr *JSONFileReWriter) Close() error {
	return rr.file.Close()
}
