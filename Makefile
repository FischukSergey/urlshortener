# устанавливаем переменные среды окружения
ipAddr:=localhost:8080
envRunAddr:=SERVER_ADDRESS=$(ipAddr)
envBaseURL:=BASE_URL=http://$(ipAddr)
envFlagFileStoragePath:=FILE_STORAGE_PATH="./tmp/short-url-db.json"
envDatabaseDSN:=DATABASE_DSN="user=postgres password=postgres host=localhost port=5432 dbname=urlshortdb sslmode=disable"
envEnableHTTPS:=ENABLE_HTTPS=true
envTrustedSubnet:=TRUSTED_SUBNET="" #"192.168.1.0/24"
envGRPC:=ENABLE_GRPC=true

server:
				@echo "Running server"
				$(envRunAddr) $(envBaseURL) $(envFlagFileStoragePath) go run ./cmd/shortener/main.go
.PHONY: server

server-https:
				@echo "Running server"
				$(envRunAddr) $(envBaseURL) $(envDatabaseDSN) $(envEnableHTTPS) go run ./cmd/shortener/main.go
.PHONY: server-https

db:
				@echo "Running server"
				$(envRunAddr) $(envBaseURL) $(envDatabaseDSN) $(envTrustedSubnet) go run ./cmd/shortener/main.go
.PHONY: db

grpc:
				@echo "Running server"
				$(envRunAddr) $(envBaseURL) $(envDatabaseDSN) $(envTrustedSubnet) $(envGRPC) go run ./cmd/shortener/main.go
.PHONY: grpc

map:
				@echo "Running server"
				$(envRunAddr) $(envBaseURL) go run ./cmd/shortener/main.go
.PHONY: map

defaultserver:
				@echo "Running default server "
				go run ./cmd/shortener/main.go

test:
				@echo "Running unit tests"
				go test -race -count=1 -cover ./...
				#go test ./internal/app/handlers/geturl/
				#go test ./internal/app/handlers/saveurl/
.PHONY: test

autotest:
				@echo "Runing autotest"
				go build -o ./cmd/shortener/shortener ./cmd/shortener/*.go
				
				/Users/sergeymac/dev/urlshortener/shortenertestbeta-darwin-arm64 -test.v -test.run=^TestIteration15$ \
				-binary-path=cmd/shortener/shortener \
				-file-storage-path=tmp/short-url-db.json \
				-source-path=./ \
				-server-port=localhost:8080 \
				-database-dsn='user=postgres password=postgres host=localhost port=5432 dbname=urlshortdb sslmode=disable'
.PHONY: autotest

testcover:
				@echo "Running unit tests into file"
				go test -coverpkg=./internal/app/... -coverprofile=coverage.out -covermode=count ./internal/app/...
#				go test -coverprofile=coverage.out ./... 
				go tool cover -func=coverage.out
.PHONY: testcover

my-lint:
				@echo "Running lint"
				go build -o ./cmd/staticlint/mylint ./cmd/staticlint/*.go
				./cmd/staticlint/mylint ./... 2> ./cmd/staticlint/result.txt
.PHONY: my-lint

clear-my-lint:
				@echo "Clearing lint"
				rm -rf ./cmd/staticlint/result.txt
.PHONY: clear-my-lint

proto:
				@echo "Generating proto"
				protoc --go_out=. --go_opt=paths=source_relative \
				--go-grpc_out=. --go-grpc_opt=paths=source_relative \
				internal/proto/contracts.proto
.PHONY: proto

# curl -v -X GET 'http://localhost:8080/map'
# curl -v -d "http://yandex.ru" -X POST 'http://localhost:8080/'
# curl -v -d '{"url": "https://codewars.com"}' -H "Content-Type: application/json" POST 'http://localhost:8080/api/shorten'
# curl -v -X GET 'http://localhost:8080/map' -H "Accept-Encoding: gzip"
# /Users/sergeymac/dev/urlshortener/shortenertestbeta-darwin-arm64 -test.v -test.run=^TestIteration9$ -binary-path=cmd/shortener/shortener -file-storage-path=tmp/short-url-db.json -source-path=tmp/short-url-db.json -database-dsn=urlshortdb
# pg_ctl -D /usr/local/pgsql/data stop/start
# go build -o shortener *.go

# go build -o ./cmd/staticlint/mylint ./cmd/staticlint/main.go
#./cmd/staticlint/mylint ./... 2> ./cmd/staticlint/result.txt

# goimports -local "github.com/FischukSergey/urlshortener.git" -w -l ./
# curl -Lv --cacert server.crt https://localhost:8080 /проверка самоподписанного сертификата	
# curl -Lv --key server.key --cert server.crt https://localhost:8080 /проверка подписанного сертификата
