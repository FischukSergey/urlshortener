#Устанавливаем переменную среды, где находится файл настроек local.yaml
#path:=CONFIG_PATH=./config/local.yaml

.PHONY: server
server:
				@echo "Running server"
				go run ./cmd/shortener/main.go
#				$(path) go run ./cmd/myrestapi/main.go
#				open http://localhost:8082/

.PHONY: test
test:
				@echo "Running unit tests"
				go test ./internal/app/handlers/geturl/
				go test ./internal/app/handlers/saveurl/
#				cd ./internal/app/handlers/geturl/