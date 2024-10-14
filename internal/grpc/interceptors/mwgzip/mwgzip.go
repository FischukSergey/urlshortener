//так не работает, сжимать только тело запроса/ответа не получается
//сжимать можно весь пакет целиком через  установку и инициализацию grpc.gzip пакета
package mwgzip

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryGzipInterceptor gzip interceptor
func UnaryGzipInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(status.Code(nil), "failed to get metadata")
	}
	slog.Info("request metadata", "metadata", md)
	//если заголовок Content-Encoding содержит gzip, то распаковываем тело запроса
	if mdContainGzip(md["content-encoding"]) {
		slog.Info("decompressing request")
		decompressedReq, err := decompressRequest(req)
		if err != nil {
			slog.Error("failed to decompress request body", "error", err)
			return nil, status.Errorf(status.Code(err), "failed to decompress request body")
		}
		req = decompressedReq
	}
	//обрабатываем запрос
	resp, err := handler(ctx, req) //получаем ответ от следующего обработчика
	if err != nil {
		return nil, status.Errorf(status.Code(err), "failed to process request")	
	}
	//если заголовок запроса содержал Accept-Encoding и в нем был gzip, то сжимаем тело ответа
	if mdContainGzip(md["accept-encoding"]) {
		slog.Info("compressing response")
		compressedResp, err := compressResponse(resp)
		if err != nil {
			slog.Error("failed to compress response body", "error", err)
			return nil, status.Errorf(status.Code(err), "failed to compress response body")
		}
		resp = compressedResp
	}
	return resp, nil
}

func compressResponse(resp any) ([]byte, error) {
	respBytes, ok := resp.([]byte)
	if !ok {
		return nil, status.Errorf(status.Code(nil), "failed to convert response to bytes")
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer func() {
		if err := gz.Close(); err != nil {
			slog.Error("failed to close gzip writer", "error", err)
		}
	}()
	_, err := gz.Write(respBytes)
	if err != nil {
		return nil, status.Errorf(status.Code(err), "failed to write to gzip writer")
	}
	return buf.Bytes(), nil
}

func mdContainGzip(s []string) bool {
	for _, v := range s {
		if strings.Contains(v, "gzip") {
			return true
		}
	}
	return false
}

// decompressRequest распаковывает тело запроса
func decompressRequest(req interface{}) ([]byte, error) {
	reqBytes, ok := req.([]byte)
	if !ok {
		return nil, status.Errorf(status.Code(nil), "failed to convert request to bytes")
	}

	gz, err := gzip.NewReader(bytes.NewReader(reqBytes))
	if err != nil {
		return nil, status.Errorf(status.Code(err), "failed to create gzip reader")
	}
	defer func() {
		if err := gz.Close(); err != nil {
			slog.Error("failed to close gzip reader", "error", err)
		}
	}()

	decompressedReq, err := io.ReadAll(gz)
	if err != nil {
		return nil, status.Errorf(status.Code(err), "failed to decompress request body")
	}
	return decompressedReq, nil
}
