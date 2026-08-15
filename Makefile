include .env
export

run:
	go run ./cmd/server/
migrate-up:
	@goose -dir migrations postgres "$(DATABASE_URL)" up
migrate-down:
	@goose -dir migrations postgres "$(DATABASE_URL)" down
db-up:
	docker compose up -d
db-down:
	docker compose down
db-shell:
	@docker compose exec db psql -U delivery_user -d delivery
fmt:
	go fmt ./...
lint:
	golangci-lint run
vet:
	go vet ./...
check: fmt vet lint