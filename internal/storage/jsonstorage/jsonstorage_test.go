package jsonstorage

import (
	"os"
	"strings"
	"testing"

	"github.com/FischukSergey/urlshortener.git/config"
)

// тест на заполнение мапы из json файла
func TestReadToMap(t *testing.T) {
	// создаем временный файл с JSON content
	file, err := os.CreateTemp("", "testfile*.json")
	if err != nil {
		t.Fatalf("Не удалось создать временный файл: %v", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Fatalf("ошибка удаления временного файла: %v", err)
		}
	}()

	content := `{"short_url":"short1","original_url":"http://example.com","uuid":"1","delete_flag":false}
{"short_url":"short2","original_url":"http://example.org","uuid":"2","delete_flag":true}`
	if _, err := file.Write([]byte(content)); err != nil {
		t.Fatalf("ошибка записи в временный файл: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("ошибка закрытия временного файла: %v", err)
	}

	// инициализируем JSONFileReader
	fr, err := NewJSONFileReader(file.Name())
	if err != nil {
		t.Fatalf("ошибка инициализации JSONFileReader: %v", err)
	}

	// создаем мапу для хранения результатов
	mapURL := make(map[string]config.URLWithUserID)

	// вызываем ReadToMap
	if err := fr.ReadToMap(mapURL); err != nil {
		t.Fatalf("ReadToMap() error = %v", err)
	}

	// проверяем результаты
	if len(mapURL) != 2 {
		t.Errorf("Expected 2 entries in map, got %d", len(mapURL))
	}

	expected := map[string]config.URLWithUserID{
		"short1": {OriginalURL: "http://example.com", UserID: 1, DeleteFlag: false},
		"short2": {OriginalURL: "http://example.org", UserID: 2, DeleteFlag: true},
	}

	for k, v := range expected {
		if got, ok := mapURL[k]; !ok || got != v {
			t.Errorf("Expected mapURL[%s] = %v, got %v", k, v, got)
		}
	}
}

// тест на дублирование shortURL при заполнении мапы	из json файла
func TestReadToMap_DuplicateShortURL(t *testing.T) {
	// создаем временный файл с JSON content
	file, err := os.CreateTemp("", "testfile*.json")
	if err != nil {
		t.Fatalf("Не удалось создать временный файл: %v", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Fatalf("ошибка удаления временного файла: %v", err)
		}
	}()

	content := `{"short_url":"short1","original_url":"http://example.com","uuid":"1","delete_flag":false}
{"short_url":"short1","original_url":"http://example.org","uuid":"2","delete_flag":true}`
	if _, err := file.Write([]byte(content)); err != nil {
		t.Fatalf("ошибка записи в временный файл: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("ошибка закрытия временного файла: %v", err)
	}

	// инициализируем JSONFileReader
	fr, err := NewJSONFileReader(file.Name())
	if err != nil {
		t.Fatalf("ошибка инициализации JSONFileReader: %v", err)
	}

	// создаем мапу для хранения результатов
	mapURL := make(map[string]config.URLWithUserID)

	// вызываем ReadToMap
	if err := fr.ReadToMap(mapURL); err != nil {
		t.Fatalf("ReadToMap() error = %v", err)
	}

	// проверяем результаты
	if len(mapURL) != 1 {
		t.Errorf("Expected 1 entry in map, got %d", len(mapURL))
	}
}

// тест на запись в json файл
func TestWrite(t *testing.T) {
	// создаем временный файл с JSON content
	file, err := os.CreateTemp("", "testfile*.json")
	if err != nil {
		t.Fatalf("Не удалось создать временный файл: %v", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Fatalf("ошибка удаления временного файла: %v", err)
		}
	}()

	// инициализируем JSONFileWriter
	fw, err := NewJSONFileWriter(file.Name())
	if err != nil {
		t.Fatalf("ошибка инициализации JSONFileWriter: %v", err)
	}

	// создаем структуру для записи
	raw := config.SaveShortURL{
		ShortURL:    "short1",
		OriginalURL: "http://example.com",
		UserID:      1,
	}

	// вызываем Write
	if err := fw.Write(raw); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// проверяем результаты
	if err := fw.Close(); err != nil {
		t.Fatalf("ошибка закрытия файла: %v", err)
	}

	// считываем содержимое файла
	content, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatalf("ошибка чтения файла: %v", err)
	}

	// проверяем, что файл содержит запись
	if !strings.Contains(string(content), "short1") {
		t.Errorf("Expected file to contain 'short1', got %s", string(content))
	}
}

// тест на запись флага delete_flag
func TestWrite_DeleteFlag(t *testing.T) {
	// создаем временный файл с JSON content
	file, err := os.CreateTemp("", "testfile*.json")
	if err != nil {
		t.Fatalf("Не удалось создать временный файл: %v", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Fatalf("ошибка удаления временного файла: %v", err)
		}
	}()

	// инициализируем JSONFileWriter
	fw, err := NewJSONFileReWriter(file.Name())
	if err != nil {
		t.Fatalf("ошибка инициализации JSONFileWriter: %v", err)
	}

	// создаем структуру для записи
	raw := config.URLWithUserID{
		OriginalURL: "http://example.com",
		UserID:      1,
		DeleteFlag:  true,
	}
	// вызываем 	DeleteFlag
	if err := fw.DeleteFlag(map[string]config.URLWithUserID{"short1": raw}); err != nil {
		t.Fatalf("DeleteFlag() error = %v", err)
	}

	// проверяем результаты
	if err := fw.Close(); err != nil {
		t.Fatalf("ошибка закрытия файла: %v", err)
	}

	// считываем содержимое файла
	content, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatalf("ошибка чтения файла: %v", err)
	}

	// проверяем, что файл содержит запись
	if !strings.Contains(string(content), "delete_flag") {
		t.Errorf("Expected file to contain 'delete_flag: true', got %s", string(content))
	}
}

// тест на закрытие файла
func TestClose(t *testing.T) {
	file, err := os.CreateTemp("", "testfile*.json")
	if err != nil {
		t.Fatalf("Не удалось создать временный файл: %v", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Fatalf("ошибка удаления временного файла: %v", err)
		}
	}()

	fw, err := NewJSONFileWriter(file.Name())
	if err != nil {
		t.Fatalf("ошибка инициализации JSONFileWriter: %v", err)
	}
	// вызываем Close
	if err := fw.Close(); err != nil {
		t.Fatalf("ошибка закрытия файла: %v", err)
	}
}
