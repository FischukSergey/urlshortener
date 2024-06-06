package geturl

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetURL(t *testing.T) {
	type want struct {
		statusCode  int
		location    string
		resError    string
	}

	tests := []struct {
		name  string
		alias string
		want  want
	}{
		{
			name:  "simple test",
			alias: "practicum",
			want: want{
				statusCode:  307,
				location:    "https://practicum.yandex.ru/",
				resError:    "",
			},
		},
		{
			name:  "alias is empty",
			alias: "",
			want: want{
				statusCode:  404,
				resError:    "404 page not found\n",
			},
		},
		{
			name:  "alias not found",
			alias: "dsfghjjkj",
			want: want{
				statusCode:  400,
				resError:    "alias not found\n",
			},
		},
	}

	var log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aliasGet := fmt.Sprintf("/%s", tt.alias)
			fmt.Println(aliasGet)

			r := chi.NewRouter()
			r.Get("/{alias}", GetURL(log))

			request, err := http.NewRequest(http.MethodGet, aliasGet, nil) 
			require.NoError(t, err)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.location, w.Header().Get("Location"))

			result, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.resError, string(result))
		})
	}
}
