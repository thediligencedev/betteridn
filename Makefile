BINARY_NAME=server
MIGRATIONS_DIR=./cmd/migrations
POSTGRES_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
BUILD_DIR=bin

include .env
export

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run:
	go run ./cmd/server

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database $(POSTGRES_URL) up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database $(POSTGRES_URL) down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -seq -ext sql -dir $(MIGRATIONS_DIR) $$name

migrate-force:
	@read -p "Enter migration version to force: " version; \
	migrate -path $(MIGRATIONS_DIR) -database $(POSTGRES_URL) force $$version

migrate-reset:
	migrate -path $(MIGRATIONS_DIR) -database $(POSTGRES_URL) down
	migrate -path $(MIGRATIONS_DIR) -database $(POSTGRES_URL) force 0
	migrate -path $(MIGRATIONS_DIR) -database $(POSTGRES_URL) up

clean:
	rm -f $(BINARY_NAME)

test:
	go test -v ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

.PHONY: build run migrate-up migrate-down migrate-create migrate-force migrate-reset clean test fmt vet
