package stats

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/logger"
)

// StatsGetter интерфейс для получения статистики
type StatsGetter interface {
	GetStats(ctx context.Context) (config.Stats, error)
}

// GetStats хендлер запроса статистики, возвращает количество URL и пользователей
// Доступно только для пользователя подсети, указанной в конфиге
func GetStats(log *slog.Logger, storage StatsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Handler: GetStats")
		w.Header().Set("Content-Type", "application/json")

		//получение статистики из хранилища
		stats, err := storage.GetStats(r.Context())
		log.Debug("статистика получена", slog.Any("stats", stats))
		if err != nil {
			log.Error("Ошибка при получении статистики", logger.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "Ошибка при получении статистики"})
			return
		}
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, stats)
	}
}
