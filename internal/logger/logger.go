package logger

import (
	"log/slog"
)

// Err логирование ошибок
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
