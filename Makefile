run: build
	@./bin/app

build:
	@go build -o bin/app .

migrate:
	@go run cmd/migrate/main.go
