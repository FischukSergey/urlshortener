package saveurl

import (
	"context"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/go-chi/chi/middleware"
)

type URLSaver interface {
	SaveStorageURL(ctx context.Context, alias, URL string) error
	GetStorageURL(alias string) (string, bool)
}

const aliasLength = 8 //для генератора случайного алиаса

// PostURL хендлер добавления (POST) сокращенного URL
func PostURL(log *slog.Logger, storage URLSaver) http.HandlerFunc {
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
		if _, err := url.ParseRequestURI(string(body)); err != nil {
			http.Error(w, "Invalid request URL", http.StatusBadRequest)
			log.Error("Invalid request URL")
			return
		}

		alias = NewRandomString(aliasLength) //генерируем произвольный алиас длины {aliasLength}

		if _, ok := storage.GetStorageURL(alias); ok {
			http.Error(w, "alias already exist", http.StatusConflict)
			log.Error("Can't add, alias already exist", slog.String("alias:", alias))
			return
		}

		ctx := context.Background()
		_ = storage.SaveStorageURL(ctx, alias, string(body))

		msg = append(msg, config.FlagBaseURL)
		msg = append(msg, alias)
		newPath = strings.Join(msg, "/")

		// fmt.Println(UrlStorage) //отладка убрать

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(newPath))
		log.Info("Request POST successful",
			slog.String("alias", alias),
			slog.String("IDrequest", middleware.GetReqID(r.Context())),
		)
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
