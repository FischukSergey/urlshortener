package gzipper

import (
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/geturl"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/saveurl"
	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/saveurljson"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMwGzipper(t *testing.T) {
	type want struct {
		gzipEncode string
		statusCode int
	}
	tests := []struct {
		name        string
		contentType string
		httpMethod  string
		uriString   string
		bodyGzip    string
		want        want
	}{
		{
			name:        "test POST JSON",
			contentType: "application/json",
			httpMethod:  "POST",
			uriString:   "/api/shorten",
			bodyGzip:    `{"url":"https://practicum.yandex.ru/"}`,
			want: want{
				gzipEncode: `{"result":`,
				statusCode: 201,
			},
		},
		{
			name:        "test POST TEXT",
			contentType: "text/plain",
			httpMethod:  "POST",
			uriString:   "/",
			bodyGzip:    "https://practicum.yandex.ru/",
			want: want{
				gzipEncode: config.FlagBaseURL,
				statusCode: 201,
			},
		},
		{
			name:        `test GET\{alias}`,
			contentType: "text/plain",
			httpMethod:  "GET",
			uriString:   "/practicum",
			want: want{
				gzipEncode: "https://practicum.yandex.ru/",
				statusCode: 307,
			},
		},
	}

	var log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	var m = mapstorage.NewMap()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//кодируем запрос
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)
			_, err := zw.Write([]byte(tt.bodyGzip))
			require.NoError(t, err)
			err = zw.Close()
			require.NoError(t, err)

			//подключаем middleware c gzip
			r := chi.NewRouter()
			r.Use(NewMwGzipper(log))
			//определяем endPoint
			r.Post("/api/shorten", saveurljson.PostURLjson(log, m))
			r.Post("/", saveurl.PostURL(log, m))
			r.Get("/{alias}", geturl.GetURL(log, m))

			//запускаем сервер
			request := httptest.NewRequest(tt.httpMethod, tt.uriString, bytes.NewReader(buf.Bytes()))
			request.Header.Set("Content-Type", tt.contentType)
			request.Header.Add("Content-Encoding", "gzip")
			request.Header.Add("Accept-Encoding", "gzip")

			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			//читаем ответ сервера
			result := w.Result()
			gz, err := gzip.NewReader(result.Body)
			require.NoError(t, err)
			defer gz.Close()

			res, err := io.ReadAll(gz)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			//сравниваем ответ с эталоном

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			switch tt.httpMethod {
			case "POST":
				assert.Contains(t, string(res), tt.want.gzipEncode)
			case "GET":
				assert.Equal(t, result.Header.Get("Location"), tt.want.gzipEncode)
			}

		})
	}
}
