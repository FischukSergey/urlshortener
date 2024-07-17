package getuserallurl

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/FischukSergey/urlshortener.git/config"
	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/go-chi/render"
)

type AllURLGetter interface { //интерфейс с методом поиска по хранилищу (только для БД)
	GetAllUserURL(ctx context.Context, userId int) ([]AllURLUserID, error)
}

type AllURLUserID struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	//Error       string `json:"error,omitempty"`
}

// GetUserAllURL хендлер запроса всех записей пользователя полного и сокращенного URL
// Пользователь определяется из токена в куки
func GetUserAllURL(log *slog.Logger, storage AllURLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Handler: GetUserAllURL")

		id := r.Context().Value(auth.CtxKeyUser).(int)

		fmt.Println(id)
		if id <= 0 { //ID проверяется в хендлере, здесь на всякий случай
			log.Error("bad request, id user absent")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		//TODO: вставить вызов метода стораджа ДБ
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
		for i, resp:=range result{ //готовим нужный формат для ответа
			result[i]=AllURLUserID{
				ShortURL: config.FlagBaseURL + "/" + resp.ShortURL,
			}
		}
		render.JSON(w, r, result) //отправляем json ответ

		log.Info("GET all URL for ID user success", slog.String("user ID:", strconv.Itoa(id)))
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
	}
}
