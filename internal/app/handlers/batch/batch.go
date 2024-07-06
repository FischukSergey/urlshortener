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
	SaveStorageURL(ctx context.Context, alias, URL string) error
	GetStorageURL(ctx context.Context, alias string) (string, bool)
}

type Request struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type Response struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
	Error         string `json:"error,omitempty"`
}

// PostBatch() хендлер обработки множественных записей json
func PostBatch(log *slog.Logger, storage BatchSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Handler: PostBatch")
		var response []Response
		var request []Request
		w.Header().Set("Content-Type", "application/json")

		err := render.DecodeJSON(r.Body, &request)

		if errors.Is(err, io.EOF) || len(request) == 0 {
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
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		for _, req := range request {
			if _, err := url.ParseRequestURI(req.OriginalURL); err != nil {
				log.Error("Invalid request URL")
				continue
			}

			alias := utilitys.NewRandomString(config.AliasLength)

			err = storage.SaveStorageURL(ctx, alias, req.OriginalURL)
			if err != nil {
				log.Error("Can't save JSON batch", err)
				return
			}
			response = append(response, Response{CorrelationID: req.CorrelationID,
				ShortURL: alias})

			render.JSON(w, r, response)

			log.Info("Request POST batch json successful", slog.String("IDrequest", middleware.GetReqID(r.Context())))

		}

	}
}
