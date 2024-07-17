package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey int

const (
	TOKEN_EXP         = time.Hour * 3
	SECRET_KEY        = "supersecretkey"
	CtxKeyUser ctxKey = iota
)

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString() (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		// собственное утверждение
		UserID: 5, //int(ctxKeyUser),
	})

	// создаём подписанную строку токена
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}
	fmt.Println(tokenString) //TODO: убрать
	// возвращаем строку токена
	return tokenString, nil
}

// проверка валидности токена
func GetUserId(tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		})
	if err != nil {
		return -1
	}

	if !token.Valid {
		//fmt.Println("Token is not valid")
		return -1
	}

	//fmt.Println("Token is valid")
	return claims.UserID
}

func NewMwToken(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		log.Info("middleware authorize started")

		Authorize := func(w http.ResponseWriter, r *http.Request) {
			var userId int
			tokenCookie, err := r.Cookie("session_token")

			if err == nil {
				userId = GetUserId(tokenCookie.Value) //проверим на валидность токен
			}
			//TODO: если нет токена в куки или он не валиден, то создаем токен BuildJWTString для методов POST
			if (userId == -1 || err != nil) && r.Method == "POST" {
				valueCookie, err := BuildJWTString() //вызываем функцию создания и прерываем обработку
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Error("can`t create signed token", err)
					return
				}

				http.SetCookie(w, &http.Cookie{ //пишем подписанную куки в ответ запросу
					Name:  "session_token",
					Value: valueCookie, //дописать
				})
				log.Info("signed token create successfully")
				return
			}
			//если метод не POST и куки не читается, возвращаем ошибку
			if err != nil && r.Method != "POST" {
				log.Error("can`t read cookie", err)
				http.Error(w, "can`t read cookie", http.StatusBadRequest)
				return
			}
			//если куки прочитался, но не валиден - возвращаем ошибку 401
			if userId < 0 {
				log.Error("bad request, id user absent")
				http.Error(w, "bad request, id user not identity",http.StatusUnauthorized)
				return
			}

			//если все успешно - пишем в контекст ID пользователя
			log.Info("token validate", slog.String("user ID:", strconv.Itoa(userId)))
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxKeyUser, userId)))
		}

		return http.HandlerFunc(Authorize)
	}
}
