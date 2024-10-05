package stats

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/models"
)

// mockStorage для тестирования
type mockStorage struct{}

// GetStats для тестирования mockStorage
func (m *mockStorage) GetStats(ctx context.Context) (config.Stats, error) {
	stats := config.Stats{
		URLs:  10,
		Users: 5,
	}
	return stats, nil
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

// TestGetStats для тестирования GetStats
func TestGetStats(t *testing.T) {
	type args struct {
		log     *slog.Logger
		storage StatsGetter
	}

	stringSubnet := "192.168.1.0/24"
	trustedSubnet, err := models.NewTrustedSubnet(stringSubnet)
	if err != nil {
		t.Fatalf("Ошибка при создании доверенной подсети: %v", err)
	}

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
			req, err := http.NewRequest("GET", "/api/internal/stats", nil)
			require.NoError(t, err)
			req.Header.Add("X-Real-IP", tt.userIP)

			rr := httptest.NewRecorder()
			handler := GetStats(log, &mockStorage{}, &trustedSubnet)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
			require.Equal(t, tt.expectedStats, rr.Body.String())
		})
	}
}
