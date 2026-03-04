
include .env

.PHONY: migrate-up migrate-down generate dev

migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

generate:
	sqlc generate

dev: migrate-up generate