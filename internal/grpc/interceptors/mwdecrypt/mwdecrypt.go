package mwdecrypt

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/FischukSergey/urlshortener.git/internal/app/middleware/auth"
	"github.com/FischukSergey/urlshortener.git/internal/logger"
)

// ctxKey тип для ключей контекста
type ctxKey int

// CtxKeyUserGrpc константа для ключа контекста
const (
	CtxKeyUserGrpc ctxKey = iota + 1
)

var log = slog.New( //инициализируем логгер
	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
)
var (
	token  string
	userID int
	md     metadata.MD
)

// UnaryDecryptInterceptor проверка токена
func UnaryDecryptInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	//получаем токен из контекста(метаданных), проверяем на валидность и получаем ID пользователя
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("session_token")
		if len(values) > 0 {
			token = values[0]
			log.Debug("token found")
			//получаем ID пользователя из токена
			userID = auth.GetUserID(token)
			if userID == -1 {
				log.Debug("userID not found")
			} else {
				log.Debug("userID found", slog.Int("userID", userID))
			}
		} else {
			userID = -1
			log.Debug("token not found")
		}
	}
	//если токена нет и метод POST генерируем новый токен и записываем в metadata ответа
	if userID == -1 && strings.Contains(info.FullMethod, "Post") { //если токена нет и метод POST
		newToken, id, err := auth.BuildJWTString() //вызываем функцию создания и прерываем обработку если неудача
		if err != nil {
			log.Error("can`t create signed token", logger.Err(err))
			return nil, status.Error(codes.Internal, "can`t create signed token")
		}

		// записываем newToken в response metadata
		err = grpc.SetHeader(ctx, metadata.Pairs("session_token", newToken))
		if err != nil {
			log.Error("can`t set header", logger.Err(err))
			return nil, status.Error(codes.Internal, "can`t set token to metadata response")
		}
		//ctx = metadata.AppendToOutgoingContext(ctx, "session_token", newToken)

		userID = id //присваиваем новый ID
		log.Debug("signed token create successfully", slog.Int("ID", id))
	}
	// записываем userID в контекст
	ctx = context.WithValue(ctx, CtxKeyUserGrpc, userID)
	return handler(ctx, req)
}
