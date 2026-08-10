include .env
export

run:
	go run main.go
migrate-up:
	@goose -dir migrations postgres "$(DATABASE_URL)" up
migrate-down:
	@goose -dir migrations postgres "$(DATABASE_URL)" down
db-up:
	docker compose up -d
db-down:
	docker compose down
