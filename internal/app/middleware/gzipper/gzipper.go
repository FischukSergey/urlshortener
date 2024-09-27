package gzipper

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/FischukSergey/urlshortener.git/internal/logger"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// newCompressWriter создаёт новый compressWriter
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header возвращает http.Header
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает данные в gzip.Writer
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader устанавливает заголовок ответа
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 || statusCode == 307 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader создаёт новый compressReader
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read считывает данные из gzip.Reader
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает gzip.Reader
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// NewMwGzipper создаёт новый middleware для сжатия и декомпрессии данных
func NewMwGzipper(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		log.Info("middleware gzip encode/decode started")

		GzipFn := func(w http.ResponseWriter, r *http.Request) {
			ow := w
			// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := newCompressWriter(w)

				log.Info("body response coded",
					slog.String("Accept-Encoding", acceptEncoding))

				ow = cw //меняем стандартный Response

				defer func() {
					err := cw.Close()
					if err != nil {
						log.Error("Error close compressWriter", logger.Err(err))
					}
				}()
			}
			// проверяем, что клиент отправил серверу сжатые данные в формате gzip
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
				cr, err := newCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// меняем тело запроса на новое
				r.Body = cr

				log.Info("body request encoded",
					slog.String("uri", r.RequestURI),
					slog.String("Content-Encoding", contentEncoding))

				defer func() {
					err := cr.Close()
					if err != nil {
						log.Error("Error close compressReader", logger.Err(err))
					}
				}()
			}

			next.ServeHTTP(ow, r)
		}

		return http.HandlerFunc(GzipFn)
	}
}
