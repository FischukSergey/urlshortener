package gzipper

import (
	"bytes"
	"compress/gzip"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/saveurljson"
	"github.com/FischukSergey/urlshortener.git/internal/storage/mapstorage"
)

func TestNewMwGzipper(t *testing.T) {
	type want struct {
		gzipEncode string
	}
	type requestJSON struct {
		jsonString string
	}
	tests := []struct {
		name     string
		bodyGzip requestJSON
		want     want
	}{
		{
			name: "simple test",
			bodyGzip: requestJSON{
				jsonString: `{"url":"https://practicum.yandex.ru/"}`,
			},
			want: want{
				gzipEncode: "practicum",
			},
		},
	}

	var log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	var m = mapstorage.NewMap()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)
			_, _ = zw.Write([]byte(tt.bodyGzip.jsonString))
			_ = zw.Close()

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(buf.Bytes())) //bytes.NewReader([]byte(tt.bodyRequest.jsonString)))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Content-Encoding", "gzip")

			w := httptest.NewRecorder()

			h := http.HandlerFunc(saveurljson.PostURLjson(log, m))
			h(w, request)

		})
	}
}
