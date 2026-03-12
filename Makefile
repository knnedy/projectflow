include .env

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