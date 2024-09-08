package getuserallurl

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
)

type AllURLUserID struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	//Error       string `json:"error,omitempty"`
}

type AllURLGetter interface { //интерфейс с методом поиска по хранилищу (только для БД)
	GetAllUserURL(ctx context.Context, userID int) ([]AllURLUserID, error)
}

// GetUserAllURL хендлер запроса всех записей пользователя полного и сокращенного URL
// Пользователь определяется из токена в куки
func GetUserAllURL(log *slog.Logger, storage AllURLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Handler: GetUserAllURL")
		w.Header().Set("Content-Type", "application/json")

		id := r.Context().Value(auth.CtxKeyUser).(int)

		if id == -1 { //нет ID или не валидный куки
			log.Error("bad request, not id user")
			http.Error(w, "you haven`t id user", http.StatusUnauthorized)
			return
		}

		var result []AllURLUserID
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		result, err := storage.GetAllUserURL(ctx, id)
		if err != nil {
			http.Error(w, "sql unexpected error", http.StatusInternalServerError)
			log.Error("sql unexpected error ", slog.String("ID user:", strconv.Itoa(id)))
			return
		}
		if len(result) == 0 { //если для пользователя ID не нашлось сокращенных URL
			http.Error(w, "you haven`t short urls", http.StatusNoContent)
			log.Error("user haven`t short urls ", slog.String("ID user:", strconv.Itoa(id)))
			return
		}
		//если все успешно
		for i, resp := range result { //готовим нужный формат для ответа
			result[i] = AllURLUserID{
				ShortURL:    config.FlagBaseURL + "/" + resp.ShortURL,
				OriginalURL: resp.OriginalURL,
			}
		}
		render.JSON(w, r, result) //отправляем json ответ

		log.Info("GET all URL for ID user success", slog.String("user ID:", strconv.Itoa(id)))
		w.WriteHeader(http.StatusOK)
	}
}
