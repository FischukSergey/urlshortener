package saveurljson

import (
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

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func PostURLjson(log *slog.Logger, storage saveurl.URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Handler: PostURLjson")
		var alias, newPath string
		var msg []string
		var req Request
		w.Header().Set("Content-Type", "application/json")

		err := render.DecodeJSON(r.Body, &req)

		if errors.Is(err, io.EOF) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Error("Request is empty")
			render.JSON(w, r, Response{
				Error: "empty request",
			})
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Error("failed to decode json request body", err)
			render.JSON(w, r, Response{
				Error: "failed to decode json request",
			})
			return
		}

		if _, err := url.ParseRequestURI(req.URL); err != nil {
			http.Error(w, "Invalid request URL", http.StatusBadRequest)
			log.Error("Invalid request URL")
			render.JSON(w, r, Response{
				Error: "invalid request URL",
			})
			return
		}

		alias = saveurl.NewRandomString(8) //поправить
		if _, ok := storage.GetStorageURL(alias); ok {
			http.Error(w, "alias already exist", http.StatusConflict)
			log.Error("Can't add, alias already exist", slog.String("alias:", alias))
			return
		}

		storage.SaveStorageURL(alias, req.URL)

		msg = append(msg, config.FlagBaseURL)
		msg = append(msg, alias)
		newPath = strings.Join(msg, "/")

		w.WriteHeader(http.StatusCreated)

		render.JSON(w, r, Response{
			Result: newPath,
		})

		log.Info("Request POST json successful", slog.String("alias:", alias))
	}
}
