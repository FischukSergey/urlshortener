package test

import (
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/FischukSergey/urlshortener.git/internal/proto"
	"github.com/FischukSergey/urlshortener.git/tests/suite"
)

// тест на отправку ссылки на сервер и получение оригинальной ссылки
func TestPostHandler(t *testing.T) {
	ctx, st := suite.New(t)
	//генерация рандомной ссылки
	originalURL := gofakeit.URL()

	// Initialize metadata
	md := metadata.New(nil)
	// отправка ссылки на сервер
	respPost, err := st.GrpcClient.PostURL(ctx, &pb.PostURLRequest{OriginalUrl: originalURL}, grpc.Header(&md))
	require.NoError(t, err)
	require.NotNil(t, respPost)

	// получение короткой ссылки
	shortURL := respPost.GetBaseUrl()
	require.NotEmpty(t, shortURL)

	// получаем токен из заголовка ответа
	sessionToken := md.Get("session_token")
	require.NotEmpty(t, sessionToken)

	//проверка получения оригинальной ссылки по alias GetURL запросом
	short, err := url.Parse(shortURL)
	//выделяем alias из ответа
	alias := strings.TrimPrefix(short.Path, "/")

	require.NoError(t, err)
	require.NotNil(t, short)

	//получение оригинальной ссылки
	respGet, err := st.GrpcClient.GetURL(ctx, &pb.GetURLRequest{ShortUrl: alias})
	require.NoError(t, err)
	require.NotNil(t, respGet)
	require.Equal(t, originalURL, respGet.GetOriginalUrl())
}

// тест на отправку группы ссылок на сервер и получение оригинальных ссылок
func TestPostBatchHandler(t *testing.T) {
	ctx, st := suite.New(t)

	// Initialize metadata
	md := metadata.New(nil)

	//генерация 10 рандомных ссылок
	requests := make([]*pb.Request, 0, 10)
	originalURLs := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		originalURLs = append(originalURLs, gofakeit.URL())
		requests = append(requests, &pb.Request{CorrelationId: gofakeit.UUID(), OriginalUrl: originalURLs[i]})
	}

	//отправка группы ссылок на сервер
	respPostBatch, err := st.GrpcClient.PostBatch(ctx, &pb.PostBatchRequest{Requests: requests}, grpc.Header(&md))
	require.NoError(t, err)
	require.NotNil(t, respPostBatch)

	//получение токена из заголовка ответа
	sessionToken := md.Get("session_token")
	require.NotEmpty(t, sessionToken)

	//получение оригинальных ссылок
	//присвоим токен в контекст для получения ссылок по ID
	ctx = metadata.AppendToOutgoingContext(ctx, "session_token", sessionToken[0])

	//получение оригинальных ссылок по ID
	respGetBatch, err := st.GrpcClient.GetUserURLs(ctx, &pb.GetUserURLsRequest{})
	require.NoError(t, err)
	require.NotNil(t, respGetBatch)

	//проверка соответствия оригинальных ссылок
	getOriginalURLs := make([]string, 0, 10)
	for _, resp := range respGetBatch.GetUrls() {
		getOriginalURLs = append(getOriginalURLs, resp.GetOriginalUrl())
	}
	sort.Strings(originalURLs)
	sort.Strings(getOriginalURLs)
	require.Equal(t, originalURLs, getOriginalURLs)
}
