package saveurljson

import (
	"bytes"
	"context"
	"io"
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

func TestPostURLjson(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		respJSON    string
	}
	type requestJSON struct {
		jsonString string
	}

	tests := []struct {
		name        string
		bodyRequest requestJSON
		want        want
	}{
		{
			name: "simple test",
			bodyRequest: requestJSON{
				jsonString: `{"url":"https://practicum.yandex.ru/"}`,
			},
			want: want{
				contentType: "application/json",
				statusCode:  201,
			},
		},
		{
			name:        "test '' bodyRequest",
			bodyRequest: requestJSON{""},
			want: want{
				contentType: "application/json",
				statusCode:  400,
				respJSON:    `{"error":"empty request"}` + "\n",
			},
		},
		{
			name:        "test bad url",
			bodyRequest: requestJSON{jsonString: `{"URLjson":"99tp://practicum.yandex.ru/"}`},
			want: want{
				contentType: "application/json",
				statusCode:  400,
				respJSON:    `{"error":"invalid request URL"}` + "\n",
			},
		},

		// TODO: добавить проверку на существующий алиас когда будет настоящий, а не произвольный.
		// TODO: заменить проверку на валидность jsonString
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
			requestTest := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader([]byte(tt.bodyRequest.jsonString)))
			request := requestTest.WithContext(context.WithValue(requestTest.Context(), auth.CtxKeyUser, 5))

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
