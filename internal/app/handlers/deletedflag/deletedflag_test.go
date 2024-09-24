package deletedflag

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	//"time"

	"github.com/stretchr/testify/assert"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
)

func TestDeleteShortURL(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	var m = mapstorage.NewMap()
	m.URLStorage["abc123"] = config.URLWithUserID{
		OriginalURL: "https://practicum.yandex.ru/",
		UserID:      1,
		DeleteFlag:  false,
	}
	m.URLStorage["def456"] = config.URLWithUserID{
		OriginalURL: "https://golangify.com/map",
		UserID:      1,
		DeleteFlag:  false,
	}
	type want struct {
		mapStorage map[string]config.URLWithUserID
		statusCode int
	}

	tests := []struct {
		name        string
		want        want
		requestBody []string
		userID      int
	}{
		{
			name:        "Valid request",
			userID:      1,
			requestBody: []string{"abc123", "def456"},
			want: want{
				statusCode: http.StatusAccepted,
				mapStorage: map[string]config.URLWithUserID{
					"abc123": {
						OriginalURL: "https://practicum.yandex.ru/",
						UserID:      1,
						DeleteFlag:  true,
					},
					"def456": {
						OriginalURL: "https://golangify.com/map",
						UserID:      1,
						DeleteFlag:  true,
					},
				},
			},
		},
		{
			name:        "Empty request",
			userID:      1,
			requestBody: []string{},
			want: want{
				statusCode: http.StatusAccepted,
			},
		},
		{
			name:        "Invalid JSON",
			userID:      1,
			requestBody: nil,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := DeleteShortURL(log, m.DelChan)

			var body []byte
			var err error
			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			} else {
				body = []byte("{invalid json}")
			}

			req, err := http.NewRequest("DELETE", "/api/user/urls", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			ctx := context.WithValue(req.Context(), auth.CtxKeyUser, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.want.statusCode {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tt.want.statusCode)
			}

			if tt.want.statusCode == http.StatusAccepted {
				//time.Sleep(1 * time.Second) // ожидание отправки сообщений в канал
				for _, shortURL := range tt.requestBody {
					assert.Equal(t, tt.want.mapStorage[shortURL].DeleteFlag, true)
				}
			}
		})
	}
}
