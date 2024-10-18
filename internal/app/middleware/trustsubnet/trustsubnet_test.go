package trustsubnet

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"log/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/stretchr/testify/require"

	"github.com/FischukSergey/urlshortener.git/config"
)

// mockStorage для тестирования
type mockStorage struct{}

// MockGetStats для тестирования mockStorage
func MockGetStats(log *slog.Logger, storage mockStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := config.Stats{
			URLs:  10,
			Users: 5,
		}
		w.Header().Set("Content-Type", "application/json")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, stats)
	}
}

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

// test структура для тестирования
type test struct {
	name           string
	userIP         string
	expectedStats  string
	expectedStatus int
}

// TestMwTrustSubnet для тестирования MwTrustSubnet
func TestMwTrustSubnet(t *testing.T) {
	type args struct {
		storage mockStorage
		log     *slog.Logger
	}

	flagTrustedSubnets := "192.168.1.0/24"

	tests := []test{
		{
			name:           "Valid stats",
			userIP:         "192.168.1.1",
			expectedStatus: http.StatusOK,
			expectedStats:  "{\"urls\":10,\"users\":5}\n",
		},
		{
			name:           "Invalid stats",
			userIP:         "192.168.2.1",
			expectedStatus: http.StatusForbidden,
			expectedStats:  "{\"error\":\"Пользователь не из доверенной подсети\"}\n",
		},
		{
			name:           "Invalid stats",
			userIP:         "",
			expectedStatus: http.StatusBadRequest,
			expectedStats:  "{\"error\":\"Ошибка при получении IP пользователя\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(MwTrustSubnet(log, flagTrustedSubnets))
			r.Get("/api/internal/stats", func(w http.ResponseWriter, r *http.Request) {
				MockGetStats(log, mockStorage{}).ServeHTTP(w, r)
			})

			req, err := http.NewRequest("GET", "/api/internal/stats", nil)
			require.NoError(t, err)
			req.Header.Set("X-Real-IP", tt.userIP)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
			require.Equal(t, tt.expectedStats, rr.Body.String())
		})
	}
}
