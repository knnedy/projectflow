include .env

migrate-up:
	goose -dir ./internal/sql/schema postgres $(DATABASE_URL) up

migrate-down:
	goose -dir ./internal/sql/schema postgres $(DATABASE_URL) down

migrate-reset:
	goose -dir ./internal/sql/schema postgres $(DATABASE_URL) reset

migrate-status:
	goose -dir ./internal/sql/schema postgres $(DATABASE_URL) status
