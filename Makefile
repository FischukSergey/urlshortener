# устанавливаем переменные среды окружения
ipAddr:=localhost:8080
envRunAddr:=SERVER_ADDRESS=$(ipAddr)
envBaseURL:=BASE_URL=http://$(ipAddr)

server:
				@echo "Running server"
				$(envRunAddr) $(envBaseURL) go run ./cmd/shortener/main.go
.PHONY: server

defaultserver:
				@echo "Running default server "
				go run ./cmd/shortener/main.go

test:
				@echo "Running unit tests"
				go test ./internal/app/handlers/geturl/
				go test ./internal/app/handlers/saveurl/
.PHONY: test

# curl -v -X GET 'http://localhost:8080/map'
# curl -v -d "http://yandex.ru" -X POST 'http://localhost:8080/'
# curl -v -d '{"url": "https://codewars.com"}' -H "Content-Type: application/json" POST 'http://localhost:8080/api/shorten'
# curl -v -X GET 'http://localhost:8080/map' -H "Accept-Encoding: gzip"