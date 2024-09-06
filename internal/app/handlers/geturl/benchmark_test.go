package geturl

import (
		"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
	"github.com/go-chi/chi"
)

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

func BenchmarkGetURL(b *testing.B) {
	// Create a mock storage
	mockStorage := mapstorage.NewMap()
	mockStorage.URLStorage["practicum"] = config.URLWithUserID{
		OriginalURL: "https://practicum.yandex.ru/",
		UserID:      1,
	}
	mockStorage.URLStorage["map"] = config.URLWithUserID{
		OriginalURL: "https://golangify.com/map",
		UserID:      1,
	}

	// Create the handler
	handler := GetURL(log, mockStorage)

	// Create a router
	r := chi.NewRouter()
	r.Get("/{alias}", handler)

	// Run the benchmark
	b.ResetTimer()
	b.Run("BenchmarkGetURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Create a new request for each iteration
			req, _ := http.NewRequest("GET", "/practicum", nil)
			req = req.WithContext(context.Background())

			b.StartTimer()
			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			r.ServeHTTP(rr, req)
			b.StopTimer()
		}
	})
}
