package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/FischukSergey/urlshortener.git/internal/utilitys"
)

type ctxKey int

const (
	TokenExp          = time.Hour * 3
	SecretKey         = "supersecretkey"
	CtxKeyUser ctxKey = iota
)

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID.
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString() (string, int, error) {

	id := utilitys.NewRandomID(5) //TODO: сделать генерацию ID

	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		// собственное утверждение
		UserID: id, //int(ctxKeyUser),
	})

	// создаём подписанную строку токена
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", 0, err
	}
	fmt.Println(tokenString) //TODO: убрать
	// возвращаем строку токена
	return tokenString, id, nil
}

// проверка валидности токена
func GetUserID(tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
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

// NewMwToken middleware сессионной аутентификации.
// если метод POST, то создаём токен и записываем в куки,
// если метод GET, то проверяем токен на валидность и записываем в контекст
func NewMwToken(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		log.Info("middleware authorize started")

		Authorize := func(w http.ResponseWriter, r *http.Request) {
			var userID int
			tokenCookie, err := r.Cookie("session_token")

			if err == nil { //если токен есть, проверим на валидность и получим ID
				userID = GetUserID(tokenCookie.Value)
			}

			switch {
			//если нет токена в куки или он не валиден, то создаем токен BuildJWTString для методов POST
			case (userID == -1 || err != nil) && r.Method == "POST":
				valueCookie, id, err := BuildJWTString() //вызываем функцию создания и прерываем обработку если неудача
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Error("can`t create signed token", err)
					return
				}
				fmt.Println(valueCookie)
				http.SetCookie(w, &http.Cookie{ //пишем подписанную куки в ответ запросу
					Name:  "session_token",
					Value: valueCookie,
				})
				log.Info("signed token create successfully")
				userID = id //присваиваем новый ID

			case err != nil && r.Method != "POST": //если куки не прочитался и метод не POST
				log.Error("can`t read cookie", err)
				userID = -1

			case userID == -1 && r.Method != "POST":
				log.Error("id user from cookie absent or invalid")
			}

			//если все успешно - пишем в контекст ID пользователя
			log.Info("token create or validate", slog.String("user ID:", strconv.Itoa(userID)))
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxKeyUser, userID)))
		}

		return http.HandlerFunc(Authorize)
	}
}
