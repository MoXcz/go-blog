run: build
	@./bin/app

build:
	@go build -o bin/app .

migrate:
	@go run cmd/migrate/main.go

templ:
	@templ generate --watch --proxy=http://localhost:3000
