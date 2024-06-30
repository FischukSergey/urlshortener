package saveurljson

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/saveurl"
	"github.com/go-chi/render"
)

type URLSaverJSON interface {
	SaveStorageURL(alias, URL string) error
	GetStorageURL(alias string) (string, bool)
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

		log.Debug("Handler: PostURLjson")
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

		alias = saveurl.NewRandomString(8) //поправить
		if _, ok := storage.GetStorageURL(alias); ok {
			log.Error("Can't add, alias already exist", slog.String("alias:", alias))

			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, Response{
				Error: "alias already exist",
			})
			return
		}

		err = storage.SaveStorageURL(alias, req.URL)
		if err != nil {
			log.Error("Can't save JSON", err)
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

		// render.JSON(w, r, Response{
		// 	Result: newPath,
		// })

		log.Info("Request POST json successful", slog.String("json:", string(resp)))
	}
}
