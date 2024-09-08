package saveurl

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v6"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
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

	mockStorage := MockURLSaver

	handler := PostURL(log, mockStorage)

	// Run the benchmark
	b.ResetTimer()
	b.Run("BenchmarkPostURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {

			body := []byte(gofakeit.URL())
			req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(body))
			req = req.WithContext(context.WithValue(req.Context(), auth.CtxKeyUser, 1))
			log.Info(string(body))

			b.StartTimer()
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			b.StopTimer()
		}
	})
}
