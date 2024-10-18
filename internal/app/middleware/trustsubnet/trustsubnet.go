package trustsubnet

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/go-chi/render"

	"github.com/FischukSergey/urlshortener.git/internal/logger"
	"github.com/FischukSergey/urlshortener.git/internal/models"
)

// TrustedSubnetGetter интерфейс для проверки доступа по подсети
type TrustedSubnetGetter interface {
	IsTrusted(ip net.IP) bool
}

// MwTrustSubnet middleware для проверки подсети
func MwTrustSubnet(log *slog.Logger, flagTrustedSubnets string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log.Info("middleware trust subnet started")

		TrustSubnet := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			//парсим подсеть из переменной окружения TRUSTED_SUBNET
			trustedSubnet, err := StartTrustedSubnet(log, flagTrustedSubnets)
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

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
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(TrustSubnet)
	}
}

// StartTrustedSubnet проверяет наличие подсети в переменной окружения TRUSTED_SUBNET
// если подсеть задана, то возвращает структуру TrustedSubnet
// если подсеть не задана, то возвращает ошибку
func StartTrustedSubnet(log *slog.Logger, flagTrustedSubnets string) (models.TrustedSubnet, error) {
	if flagTrustedSubnets != "" {
		trustedSubnet, err := models.NewTrustedSubnet(flagTrustedSubnets)
		if err != nil {
			log.Error("не удалось распарсить переменную окружения TRUSTED_SUBNET: %v", logger.Err(err))
		}
		return trustedSubnet, nil
	}
	return models.TrustedSubnet{}, fmt.Errorf("подсеть не задана")
}
