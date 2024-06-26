package saveurl

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.bodyRequest)))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(PostURL(log, m))
			h(w, request)

			result := w.Result()
			err := result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

		})
	}
}
