package saveurl

import (
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
)

type Request struct { //структура запроса на будущее
	URL   string
	Alias string
}

const aliasLength = 8 //для генератора случайного алиаса

var URLStorage = map[string]string{ // временное хранилище URLов
	"practicum": "https://practicum.yandex.ru/",
	"map":       "https://golangify.com/map",
}
//PostURL хандлер добавления (POST) сокращенного URL
//принимает в качестве параметров логгер, потом можно будет добавить 
//интерфейс с методом добавления записи в базу данных
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
		if string(body) == "" {
			http.Error(w, "Request is empty", http.StatusBadRequest)
			log.Error("Request is empty")
			return
		}
		alias = NewRandomString(aliasLength) //генерируем произвольный алиас длины {aliasLength}

		if _, ok := URLStorage[alias]; ok {
			http.Error(w, "alias already exist", http.StatusBadRequest)
			log.Error("Can't add, alias already exist", slog.String("alias:", alias))
			return
		}

		URLStorage[alias] = string(body)

		msg = append(msg, config.FlagBaseURL)
		msg = append(msg, alias)
		newPath = strings.Join(msg, "/")

		// fmt.Println(UrlStorage) //отладка убрать

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(newPath))
		log.Info("Request POST successful", slog.String("alias:", alias))
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
