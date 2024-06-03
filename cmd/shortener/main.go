package main

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"

	//"strconv"
	// "strings"
	"time"

	"github.com/go-chi/chi"
)

const aliasLength = 8 //для генератора случайного алиаса
type Request struct { //структура запроса на будущее
	URL   string
	Alias string
}

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)
var url = map[string]string{ // временное хранилище urlов
	"practicum": "https://practicum.yandex.ru/",
	"map":       "https://golangify.com/map",
}

func postURL(w http.ResponseWriter, r *http.Request) {

	log.Debug("Handler: postURL")
	var alias string

	var newPath string // 	инициализируем пустой алиас
	var msg []string

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error("Request bad")
		return
	}
	alias = NewRandomString(aliasLength) //генерируем произвольный алиас длины aliasLength

	if _, ok := url[alias]; ok {
		http.Error(w, "alias already exist", http.StatusBadRequest)
		log.Error("Can't add, alias already exist", slog.String("alias:", alias))
		return
	}

	url[alias] = string(body)

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
	//lenAlias := strconv.Itoa(len(newPath))

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	//w.Header().Set("Content-Length", lenAlias)
	w.Write([]byte(newPath))
	log.Info("Request POST successful: ", slog.String("alias:", alias))
}

func getURL(w http.ResponseWriter, r *http.Request) {

	log.Debug("Handler: getURL")

	alias := chi.URLParam(r, "alias")

	if alias == "" {
		log.Info("alias is empty")
		http.Error(w, "alias is empty", http.StatusBadRequest)
		return
	}

	url, ok := url[alias]
	if !ok {
		http.Error(w, "alias not found", http.StatusBadRequest)
		log.Error("alias not found", slog.String("alias: ", alias))
		return
	}

	/*
	   resp, err := json.Marshal(task)
	   if err != nil {
	       http.Error(w, err.Error(), http.StatusBadRequest)
	       return
	   }
	*/

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)

	// w.Write([]byte(url))
	log.Info("Request alias  successful: ", slog.String("alias:", alias), slog.String("url:", url))
	//http.Redirect(w, r, url, http.StatusTemporaryRedirect)

}

func main() {
	r := chi.NewRouter()

	// здесь регистрируйте ваши обработчики
	r.Get("/{alias}", getURL)
	r.Post("/", postURL)

	srv := &http.Server{
		Addr:         "localhost:8080",
		Handler:      r,
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 4 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Info("Initializing server", slog.String("address", srv.Addr))

	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
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
