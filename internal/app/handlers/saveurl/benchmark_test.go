package saveurl

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
	"github.com/brianvoe/gofakeit/v6"
)

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

func BenchmarkPostURL(b *testing.B) {

	// Create a mock storage
	MockURLSaver := mapstorage.NewMap()
	MockURLSaver.URLStorage["practicum"] = config.URLWithUserID{
		OriginalURL: "https://practicum.yandex.ru/",
	}
	MockURLSaver.URLStorage["map"] = config.URLWithUserID{
		OriginalURL: "https://golangify.com/map",
	}

	// Use the MockURLSaver directly as it's already a DataStore
	mockStorage := MockURLSaver

	// Create the handler
	handler := PostURL(log, mockStorage)

	// Run the benchmark
	b.ResetTimer()
	b.Run("BenchmarkPostURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Create a new request for each iteration
			// Create a sample request body
			body := []byte(gofakeit.URL())
			req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(body))
			req = req.WithContext(context.WithValue(req.Context(), auth.CtxKeyUser, 1))
			log.Info(string(body))
			
			b.StartTimer()
			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rr, req)
			b.StopTimer()
		}
	})
}
