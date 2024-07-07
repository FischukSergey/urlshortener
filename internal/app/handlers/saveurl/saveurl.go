package saveurl

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/utilitys"
	"github.com/go-chi/chi/middleware"
)

type URLSaver interface {
	SaveStorageURL(ctx context.Context, saveURL []config.SaveShortURL) error
	GetStorageURL(ctx context.Context, alias string) (string, bool)
}

const aliasLength = 8 //для генератора случайного алиаса

// PostURL хендлер добавления (POST) сокращенного URL
func PostURL(log *slog.Logger, storage URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Handler: PostURL")
		var saveURL []config.SaveShortURL
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

		alias = utilitys.NewRandomString(config.AliasLength) //генерируем произвольный алиас длины {aliasLength}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if _, ok := storage.GetStorageURL(ctx, alias); ok {
			http.Error(w, "alias already exist", http.StatusConflict)
			log.Error("Can't add, alias already exist", slog.String("alias:", alias))
			return
		}

		saveURL = append(saveURL, config.SaveShortURL{
			ShortURL:    alias,
			OriginalURL: string(body),
		})

		err = storage.SaveStorageURL(ctx, saveURL)
		if err != nil {
			http.Error(w, "Error write DB", http.StatusInternalServerError)
			log.Error("Error write DB", err)
			return
		}
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
