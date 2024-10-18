package stats

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FischukSergey/urlshortener.git/config"
)

type mockStorage struct {
	err   error
	stats config.Stats
}

func (m *mockStorage) GetStats(ctx context.Context) (config.Stats, error) {
	return m.stats, m.err
}

func TestGetStats(t *testing.T) {
	tests := []struct {
		name           string
		expectedBody   string
		mockStorage    mockStorage
		expectedStatus int
	}{
		{
			name: "Valid stats",
			mockStorage: mockStorage{
				stats: config.Stats{
					URLs:  10,
					Users: 5,
				},
				err: nil,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "{\"urls\":10,\"users\":5}\n",
		},
		{
			name: "Error stats",
			mockStorage: mockStorage{
				err: errors.New("error"),
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "{\"error\":\"Ошибка при получении статистики\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestTest := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			requestTest.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			log := slog.New(
				slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
			handler := GetStats(log, &test.mockStorage)
			handler.ServeHTTP(w, requestTest)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedBody, w.Body.String())
		})
	}
}
