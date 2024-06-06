package saveurl

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostURL(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		alias       string //на будущее
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

		// TODO: добавить проверку на существующий алиас когда будет настоящий, а не произвольный.
		// TODO: добавить проверку на валидность URL
	}
	var log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.bodyRequest)))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(PostURL(log))
			h(w, request)

			result:=w.Result()

			assert.Equal(t,tt.want.statusCode,result.StatusCode)

		})
	}
}
