package saveurl

import (
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const aliasLength = 8 //для генератора случайного алиаса

var URLStorage = map[string]string{ // временное хранилище urlов
	"practicum": "https://practicum.yandex.ru/",
	"map":       "https://golangify.com/map",
}

func PostURL(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Handler: PostURL")
		var alias, newPath string // 	инициализируем пустой алиас
		var msg []string

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Error("Request bad")
			return
		}
		alias = NewRandomString(aliasLength) //генерируем произвольный алиас длины {aliasLength}

		if _, ok := URLStorage[alias]; ok {
			http.Error(w, "alias already exist", http.StatusBadRequest)
			log.Error("Can't add, alias already exist", slog.String("alias:", alias))
			return
		}

		URLStorage[alias] = string(body)

		/*
			if string(body) == "https://practicum.yandex.ru/" {
				alias = "EwHXdJfB"
			} else {
				alias = NewRandomString(aliasLength) //генерируем произвольный алиас длины aliasLength
			}
		*/
		// var req Request
		// var buf bytes.Buffer
		// err:=render.DecodeJSON(r.Body, &req)
		// _, err := buf.ReadFrom(r.Body)
		/*
			if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Error("Bad request")
				return
			}

			if idTest, ok := tasks[task.ID]; ok { //проверка на существующий ключ мапы
				http.Error(w, "ID already exist", http.StatusBadRequest)
				log.Error("Can't add, ID already exist", slog.String("ID", idTest.ID))
				return
			}
		*/

		msg = append(msg, "http://localhost:8080/")
		msg = append(msg, alias)
		newPath = strings.Join(msg, "")
		
		// fmt.Println(UrlStorage) //отладка убрать 

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(newPath))
		log.Info("Request POST successful: ", slog.String("alias:", alias))
	}
}

// NewRandomString generates random string with given size.
func NewRandomString(size int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	b := make([]rune, size)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}

	return string(b)
}
