package jsonstorage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type JSONRaw struct {
	UUid        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type JSONFileWriter struct {
	// mx         *sync.Mutex
	file       *os.File
	JSONWriter *json.Encoder
}

//NewJSONFileWriter() создаем объект с открытым файлом для записи 
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
func (fs *JSONFileWriter) Write(alias, urlOriginal string) error {
	//fs.mx.Lock()
	// defer fs.mx.Unlock()

	//	TODO логику обработки json строки
	raw := JSONRaw{
		ShortURL:    alias,
		OriginalURL: urlOriginal,
		UUid:        "1", //TODO заменить на реальный ID
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
//инициализация объекта file и сканера строк к нему
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

func (fr *JSONFileReader) ReadToMap(mapURL map[string]string) (map[string]string, error) { //чтение файла в мапу до запуска сервера, поэтому работаем без mutex
	defer fr.file.Close()

	for fr.ScanRaw.Scan() {  //построчно читаем, декодируем и пишем в мапу
		data := fr.ScanRaw.Bytes() 
		mapLine := &JSONRaw{}
		err := json.Unmarshal(data, &mapLine)
		if err != nil {
			fmt.Println("Не удалось декодировать строку:", data)
			return nil,err
		}
		if _, ok := mapURL[mapLine.ShortURL]; ok {
			fmt.Printf("Алиас %s дублируется:\n", mapLine.ShortURL)
		}else{
			mapURL[mapLine.ShortURL]=mapLine.OriginalURL
		}
	}
	return mapURL, nil
}


func (fr *JSONFileReader) Close() error {
	return fr.file.Close()
}

// {"uuid":"1","short_url":"4rSPg8ap","original_url":"http://yandex.ru"}
// {"uuid":"2","short_url":"edVPg3ks","original_url":"http://ya.ru"}
// {"uuid":"3","short_url":"dG56Hqxm","original_url":"http://practicum.yandex.ru"}
/*

func TestSettings(t *testing.T) {
    fname := `settings.json`
    settings := Settings{
        Port: 3000,
        Host: `localhost`,
    }
    if err := settings.Save(fname); err != nil {
        t.Error(err)
    }
}

type CsvWriter struct {
mutex *sync.Mutex
csvWriter *csv.Writer
}

func NewCsvWriter(fileName string) (*CsvWriter, error) {
csvFile, err := os.Create(fileName)
if err != nil {
return nil, err
}
w := csv.NewWriter(csvFile)
return &CsvWriter{csvWriter:w, mutex: &sync.Mutex{}}, nil
}

func (w *CsvWriter) Write(row []string) {
w.mutex.Lock()
w.csvWriter.Write(row)
w.mutex.Unlock()
}

*/
