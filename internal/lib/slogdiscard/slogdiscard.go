// Slogdiscard имплементирует интерфейс slog.DiscardHandler.
// Все методы пустые, те логгер ничего не выводит.
// Применяется в тестах.
package slogdiscard

import (
	"context"
	"log/slog"
)

// DiscardHandler имплементирует интерфейс slog.DiscardHandler.
// Все методы пустые, те логгер ничего не выводит.
// Применяется в тестах.
type DiscardHandler struct{}

// NewDiscardLogger создаёт новый логгер, который ничего не выводит.
func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
}

// NewDiscardHandler создаёт новый обработчик, который ничего не выводит.
func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

// Enabled проверяет, нужно ли выводить сообщение с заданным уровнем.
func (h *DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	// Всегда возвращаем false, чтобы ничего не выводилось.
	return false
}

// Handle обрабатывает запись лога.
func (h *DiscardHandler) Handle(_ context.Context, _ slog.Record) error {
	// Всегда возвращаем nil, чтобы не было ошибки.
	return nil
}

// WithAttrs добавляет атрибуты к обработчику.
func (h *DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// Всегда возвращаем текущий обработчик, чтобы не было ошибки.
	return h
}

// WithGroup добавляет группу к обработчику.
func (h *DiscardHandler) WithGroup(_ string) slog.Handler {
	// Всегда возвращаем текущий обработчик, чтобы не было ошибки.
	return h
}
