package saveurljson

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostURLjson(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		respJSON    string
	}
	type RequestJSON struct {
		Url string `json:"url,omitempty"`
	}

	tests := []struct {
		name        string
		bodyRequest RequestJSON
		want        want
	}{
		{
			name: "simple test",
			bodyRequest: RequestJSON{
				Url: `{"url":"https://practicum.yandex.ru/"}`,
			},
			want: want{
				contentType: "application/json",
				statusCode:  201,
			},
		},
		{
			name:        "test '' bodyRequest",
			bodyRequest: RequestJSON{""},
			want: want{
				contentType: "application/json",
				statusCode:  400,
				respJSON:    `{"error":"empty request"}` + "\n",
			},
		},
		{
			name:        "test bad URL",
			bodyRequest: RequestJSON{Url: `{"url":"99tp://practicum.yandex.ru/"}`},
			want: want{
				contentType: "application/json",
				statusCode:  400,
				respJSON:    `{"error":"invalid request URL"}` + "\n",
			},
		},
		{
			name:        "test bad URL",
			bodyRequest: RequestJSON{Url: `{" ' "}`},
			want: want{
				contentType: "application/json",
				statusCode:  400,
				respJSON:    `{"error":"failed to decode json request"}` + "\n",
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
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader([]byte(tt.bodyRequest.Url)))
			request.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			h := http.HandlerFunc(PostURLjson(log, m))
			h(w, request)

			result := w.Result()
			res, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))
			if tt.want.respJSON != "" {
				assert.Equal(t, tt.want.respJSON, string(res))
			}

		})
	}
}
