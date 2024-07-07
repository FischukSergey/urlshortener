package batch

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/utilitys"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type BatchSaver interface {
	SaveStorageURL(ctx context.Context, saveURL []config.SaveShortURL) error
	GetStorageURL(ctx context.Context, alias string) (string, bool)
}

type Request struct { //структура запроса
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type Response struct { // структура ответа
	CorrelationID string `json:"correlation_id,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
	Error         string `json:"error,omitempty"`
}

// PostBatch() хендлер обработки множественных записей json
func PostBatch(log *slog.Logger, storage BatchSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Handler: PostBatch")
		var request []Request

		w.Header().Set("Content-Type", "application/json")

		err := render.DecodeJSON(r.Body, &request) //декодируем json body

		if errors.Is(err, io.EOF) || len(request) == 0 { //проверяем на пусто
			w.WriteHeader(http.StatusBadRequest)
			log.Error("Request is empty")
			render.JSON(w, r, Response{
				Error: "empty request",
			})
			return
		}
		if err != nil { //убеждаемся, что декодировали
			w.WriteHeader(http.StatusBadRequest)
			log.Error("failed to decode json request body", err)
			render.JSON(w, r, Response{
				Error: "failed to decode json request",
			})
			return
		}

		saveURL := make([]config.SaveShortURL, 0, len(request)) //слайс для записи в БД
		response := make([]Response, 0, len(request))           //слайс для json ответа
		//var saveURL []config.SaveShortURL
		//var response []Response

		for _, req := range request { //итерируем декодированные строки
			if _, err := url.ParseRequestURI(req.OriginalURL); err != nil { //проверяем на валидность исходный url
				log.Error("Invalid request URL")
				continue
			}

			alias := utilitys.NewRandomString(config.AliasLength) //вычисляем произвольный алиас

			saveURL = append(saveURL, config.SaveShortURL{ //готовим слайс для записи в БД
				ShortURL:    alias,
				OriginalURL: req.OriginalURL,
			})

			response = append(response, Response{ //готовим слайс для json ответа
				CorrelationID: req.CorrelationID,
				ShortURL:      config.FlagBaseURL + "/" + alias})

			log.Info("Request prepare", slog.String("short_url", alias), slog.String("correlation_id", req.CorrelationID))
		}

		if len(response) > 0 && len(saveURL) > 0 {

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err = storage.SaveStorageURL(ctx, saveURL) //пишем слайс в БД
			if err != nil {
				log.Error("Can't save JSON batch", err)
				return
			}

			w.WriteHeader(http.StatusCreated)
			render.JSON(w, r, response) //отправляем json ответ
			log.Info("Request POST batch json successful", slog.String("IDrequest", middleware.GetReqID(r.Context())))
		}

	}
}
