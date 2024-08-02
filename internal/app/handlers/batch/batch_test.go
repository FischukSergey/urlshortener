package batch

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FischukSergey/urlshortener.git/internal/app/handlers/batch/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostBatch(t *testing.T) {

	var log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	// type args struct {
	// 	log     *slog.Logger
	// 	storage BatchSaver
	// }

	type want struct {
		contentType  string
		statusCode   int
		bodyResponse string
		mockError    error
	}

	tests := []struct {
		name        string
		bodyRequest string
		want        want
	}{
		{
			name: "simple test",
			bodyRequest: `[
		{
    "correlation_id":"aaaaaa",
    "original_url":"https://codewars.com"
    },
    {
    "correlation_id":"bbbbbb",
    "original_url":"http://habr.com"
    }
		]`,
			want: want{
				contentType:  "application/json",
				statusCode:   201,
				bodyResponse: `[{"correlation_id":"aaaaaa","short_url":`,
			},
		},
		{
			name:        "empty request",
			bodyRequest: "",
			want: want{
				contentType:  "application/json",
				statusCode:   400,
				bodyResponse: "empty request",
				mockError:    errors.New("unexpected error"),
			},
		},
		{
			name:        "invalid request url",
			bodyRequest: `[{"correlation_id":"aaaaaa","short_url":"hhttpp://"}]`,
			want: want{
				contentType:  "text/plain; charset=utf-8",
				statusCode:   400,
				bodyResponse: "request failed, no valid url",
				mockError:    errors.New("unexpected error"),
			},
		},
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := PostBatch(tt.args.log, tt.args.storage); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("PostBatch() = %v, want %v", got, tt.want)
			ctrl := gomock.NewController(t)
			s := mock.NewMockBatchSaver(ctrl) //новый storage
			defer ctrl.Finish()

			request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader([]byte(tt.bodyRequest)))
			w := httptest.NewRecorder()
			switch {
			case tt.want.mockError != nil:
				s.EXPECT()
			default:
				s.EXPECT().SaveStorageURL(gomock.Any(), gomock.Any())
			}

			h := http.HandlerFunc(PostBatch(log, s))
			h(w, request)

			result := w.Result()
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			assert.Contains(t, string(body), tt.want.bodyResponse)
		})
	}
}
