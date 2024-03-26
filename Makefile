include .env

.PHONY: start
start:
	@echo "Starting the server..."
	go run ./cmd/api

migrate_up:
	@echo "Migrating..."
	migrate -source file://migrations -database $(DB_DSN) up

migrate_down:
	@echo "Rollback migrations..."
	migrate -source file://migrations -database $(DB_DSN) down
