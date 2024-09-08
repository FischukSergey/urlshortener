package saveurl

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
)

func TestPostURL(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	tests := []struct {
		name        string
		bodyRequest string
		want        want
	}{
		{
			name:        "simple test",
			bodyRequest: "https://practicum.yandex.ru/",
			want: want{
				contentType: "text/plain",
				statusCode:  201,
			},
		},
		{
			name:        "test '' bodyRequest",
			bodyRequest: "",
			want: want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
		{
			name:        "test bad URL",
			bodyRequest: "practicum.yandex.ru",
			want: want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
		// TODO: добавить проверку на существующий алиас когда будет настоящий, а не произвольный.
		// TODO: заменить проверку на валидность URL
	}
	var log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	var m = mapstorage.NewMap()
	m.URLStorage["practicum"] = config.URLWithUserID{
		OriginalURL: "https://practicum.yandex.ru/",
	}
	m.URLStorage["map"] = config.URLWithUserID{
		OriginalURL: "https://golangify.com/map",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestTest := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.bodyRequest)))
			w := httptest.NewRecorder()
			request := requestTest.WithContext(context.WithValue(requestTest.Context(), auth.CtxKeyUser, 5))
			h := http.HandlerFunc(PostURL(log, m))
			h(w, request)

			result := w.Result()
			err := result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

		})
	}
}
