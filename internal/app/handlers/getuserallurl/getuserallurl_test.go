package getuserallurl

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/FischukSergey/urlshortener.git/internal/models"
)

type mockStorage struct{}

func (m *mockStorage) GetAllUserURL(ctx context.Context, userID int) ([]models.AllURLUserID, error) {
	if userID == 1 {
		return []models.AllURLUserID{
			{ShortURL: "abc123", OriginalURL: "https://example.com"},
			{ShortURL: "def456", OriginalURL: "https://example.org"},
		}, nil
	}
	return nil, nil
}

var log = slog.New(
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)

type test struct {
	name           string
	expectedURLs   []models.AllURLUserID
	expectedStatus int
	userID         int
}

func TestGetUserAllURL(t *testing.T) {
	tests := []test{
		{
			name:           "User with URLs",
			userID:         1,
			expectedStatus: http.StatusOK,
			expectedURLs: []models.AllURLUserID{
				{ShortURL: config.FlagBaseURL + "/abc123", OriginalURL: "https://example.com"},
				{ShortURL: config.FlagBaseURL + "/def456", OriginalURL: "https://example.org"},
			},
		},
		{
			name:           "User without URLs",
			userID:         2,
			expectedStatus: http.StatusNoContent,
			expectedURLs:   nil,
		},
		{
			name:           "Unauthorized user",
			userID:         -1,
			expectedStatus: http.StatusUnauthorized,
			expectedURLs:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/user/urls", nil)
			require.NoError(t, err)

			ctx := context.WithValue(req.Context(), auth.CtxKeyUser, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := GetUserAllURL(log, &mockStorage{})

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var result []models.AllURLUserID
				err = json.Unmarshal(rr.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedURLs, result)
			}
		})
	}
}
