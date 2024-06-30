# устанавливаем переменные среды окружения
ipAddr:=localhost:8080
envRunAddr:=SERVER_ADDRESS=$(ipAddr)
envBaseURL:=BASE_URL=http://$(ipAddr)
envFlagFileStoragePath:=FILE_STORAGE_PATH="./tmp/short-url-db.json"

server:
				@echo "Running server"
				$(envRunAddr) $(envBaseURL) $(envFlagFileStoragePath) go run ./cmd/shortener/main.go
.PHONY: server

defaultserver:
				@echo "Running default server "
				go run ./cmd/shortener/main.go

test:
				@echo "Running unit tests"
				go test -race -count=1 ./...
				#go test ./internal/app/handlers/geturl/
				#go test ./internal/app/handlers/saveurl/
.PHONY: test

# curl -v -X GET 'http://localhost:8080/map'
# curl -v -d "http://yandex.ru" -X POST 'http://localhost:8080/'
# curl -v -d '{"url": "https://codewars.com"}' -H "Content-Type: application/json" POST 'http://localhost:8080/api/shorten'
# curl -v -X GET 'http://localhost:8080/map' -H "Accept-Encoding: gzip"
# /Users/sergeymac/dev/urlshortener/shortenertestbeta-darwin-arm64 -test.v -test.run=^TestIteration9$ -binary-path=cmd/shortener/shortener -file-storage-path=tmp/short-url-db.json -source-path=tmp/short-url-db.json