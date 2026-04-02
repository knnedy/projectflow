include .env
export

# variables
BINARY_NAME=api
MAIN_PATH=./cmd/api
DB_URL=$(DATABASE_URL)

# build
build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

# run
run:
	go run $(MAIN_PATH)

# build and run
dev: build
	./bin/$(BINARY_NAME)

# test
test:
	go test ./... -v

# test with coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# lint
lint:
	golangci-lint run ./...

# migrations
migrate-up:
	goose -dir ./internal/sql/schema postgres "$(DB_URL)" up

migrate-down:
	goose -dir ./internal/sql/schema postgres "$(DB_URL)" down

migrate-reset:
	goose -dir ./internal/sql/schema postgres "$(DB_URL)" reset

migrate-status:
	goose -dir ./internal/sql/schema postgres "$(DB_URL)" status

migrate-create:
	goose -dir ./internal/sql/schema create $(name) sql

# sqlc
sqlc:
	sqlc generate

# swagger
swagger:
	swag init -g $(MAIN_PATH)/main.go -o ./docs

# clean
clean:
	rm -f bin/$(BINARY_NAME)
	rm -f coverage.out

.PHONY: build run dev test test-coverage lint migrate-up migrate-down migrate-reset migrate-status migrate-create sqlc swagger clean