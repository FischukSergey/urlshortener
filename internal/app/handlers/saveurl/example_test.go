package saveurl

import (
	"context"
	"fmt"

	//"strconv"
	//"log/slog"
	"net/http"
	"net/http/httptest"

	//"os"
	"strings"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/FischukSergey/urlshortener.git/internal/lib/slogdiscard"
)

// MockURLSaver is a mock implementation of the URLSaver interface
type MockURLSaver struct {
	SaveFunc func(ctx context.Context, saveURL []config.SaveShortURL) error
	GetFunc  func(ctx context.Context, alias string) (string, bool)
}

func (m *MockURLSaver) SaveStorageURL(ctx context.Context, saveURL []config.SaveShortURL) error {
	return m.SaveFunc(ctx, saveURL)
}

func (m *MockURLSaver) GetStorageURL(ctx context.Context, alias string) (string, bool) {
	return m.GetFunc(ctx, alias)
}

// ExamplePostURL demonstrates the usage of the PostURL handler.
func ExamplePostURL() {
	// Create a logger for the example
	log := slogdiscard.NewDiscardLogger()
	config.FlagBaseURL = "http://localhost:8080" //mock переменная среды окружения для теста

	// Create a mock storage
	mockStorage := &MockURLSaver{
		SaveFunc: func(ctx context.Context, saveURL []config.SaveShortURL) error {
			return nil
		},
		GetFunc: func(ctx context.Context, alias string) (string, bool) {
			return "", false
		},
	}

	// Create a new request
	body := strings.NewReader("https://example.com")
	req := httptest.NewRequest(http.MethodPost, "/", body)

	// Set up a context with a mock user ID
	ctx := context.WithValue(req.Context(), auth.CtxKeyUser, 123)
	req = req.WithContext(ctx)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler := PostURL(log, mockStorage)
	handler.ServeHTTP(w, req)

	// Print the response
	fmt.Println("Status Code:", w.Code)
	//fmt.Println("Response Body:", w.Body.String())

	// Output:
	// Status Code: 201

}
