package stats

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"github.com/go-chi/render"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/logger"
)

// StatsGetter интерфейс для получения статистики
type StatsGetter interface {
	GetStats(ctx context.Context) (config.Stats, error)
}

// TrustedSubnetGetter интерфейс для проверки доступа по подсети
type TrustedSubnetGetter interface {
	IsTrusted(ip net.IP) bool
}

// GetStats хендлер запроса статистики, возвращает количество URL и пользователей
// Доступно только для пользователя подсети, указанной в конфиге
func GetStats(log *slog.Logger, storage StatsGetter, trustedSubnet TrustedSubnetGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Handler: GetStats")
		w.Header().Set("Content-Type", "application/json")

		//извлекаем IP пользователя из заголовка X-Real-IP
		ipStr := r.Header.Get("X-Real-IP")
		userIP := net.ParseIP(ipStr)
		if userIP == nil {
			log.Error("Ошибка при получении IP пользователя", slog.String("ip", ipStr))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Ошибка при получении IP пользователя"})
			return
		}

		//проверка на доступность IP из доверенной подсети
		if !trustedSubnet.IsTrusted(userIP) {
			log.Error("Пользователь не из доверенной подсети", slog.String("ip", ipStr))
			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, map[string]string{"error": "Пользователь не из доверенной подсети"})
			return
		}

		//получение статистики из хранилища
		stats, err := storage.GetStats(r.Context())
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
