package saveurljson

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/FischukSergey/urlshortener.git/internal/storage/dbstorage"
	"github.com/FischukSergey/urlshortener.git/internal/utilitys"
	"github.com/go-chi/render"
)

type URLSaverJSON interface {
	SaveStorageURL(ctx context.Context, saveURL []config.SaveShortURL) error
	GetStorageURL(ctx context.Context, alias string) (string, bool)
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// PostURLjson хендлер добавления (POST:/api/shorten) сокращенного URL.
// Запрос и ответ в формате JSON.
func PostURLjson(log *slog.Logger, storage URLSaverJSON) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.Context().Value(auth.CtxKeyUser).(int)

		log.Debug("Handler: PostURLjson")
		var saveURL []config.SaveShortURL
		var alias, newPath string
		var msg []string
		var req Request
		w.Header().Set("Content-Type", "application/json")

		err := render.DecodeJSON(r.Body, &req)

		if errors.Is(err, io.EOF) {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("Request is empty")
			render.JSON(w, r, Response{
				Error: "empty request",
			})
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("failed to decode json request body", err)
			render.JSON(w, r, Response{
				Error: "failed to decode json request",
			})
			return
		}

		if _, err := url.ParseRequestURI(req.URL); err != nil {
			log.Error("Invalid request URL")

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, Response{
				Error: "invalid request URL",
			})
			return
		}

		alias = utilitys.NewRandomString(config.AliasLength) //поправить

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if _, ok := storage.GetStorageURL(ctx, alias); ok {
			log.Error("Can't add, alias already exist", slog.String("alias:", alias))

			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, Response{
				Error: "alias already exist",
			})
			return
		}

		saveURL = append(saveURL, config.SaveShortURL{
			ShortURL:    alias,
			OriginalURL: req.URL,
			UserID:      id,
		})

		err = storage.SaveStorageURL(ctx, saveURL)

		//обработка ошибки вставки уже существующего url
		var res []string
		if errors.Is(err, dbstorage.ErrURLExists) {
			res = strings.Split(err.Error(), ":")

			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, Response{
				Result: config.FlagBaseURL + "/" + res[0],
			})
			log.Error("Request POST /api/shorten failed, url exists",
				slog.String("url", saveURL[0].OriginalURL),
			)
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, Response{
				Error: "can't save JSON",
			})

			log.Error("Can't save JSON", err)
			return
		}

		msg = append(msg, config.FlagBaseURL)
		msg = append(msg, alias)
		newPath = strings.Join(msg, "/")

		w.WriteHeader(http.StatusCreated)

		// Потому, что тест требует json пакет
		resp, err := json.Marshal(Response{
			Result: newPath,
		})
		if err != nil {
			log.Error("Can't make JSON", err)
		}
		w.Write(resp)

		log.Info("Request POST json successful", slog.String("json:", string(resp)))
	}
}
