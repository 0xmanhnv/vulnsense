.PHONY: run build test tidy migrate-up migrate-down migrate-create

# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Default Go command
GO_CMD=go

# Database URL for migration tool
# Make sure DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME are set in your .env file
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}"

# --- Main Application Commands ---

# run: runs the main application
run:
	@echo "?? Running the application..."
	@$(GO_CMD) run ./cmd/vulnsense

# build: builds the application binary
build:
	@echo "?? Building the application..."
	@$(GO_CMD) build -o ./bin/vulnsense ./cmd/vulnsense

# test: runs all tests
test:
	@echo "?? Running tests..."
	@$(GO_CMD) test ./...

# tidy: tidies go modules
tidy:
	@echo "?? Tidying Go modules..."
	@$(GO_CMD) mod tidy

# --- Database Migration Commands ---

# migrate-up: applies all up migrations
migrate-up:
	@echo "?? Applying database migrations..."
	@migrate -database "$(DB_URL)" -path migrations up

# migrate-down: reverts the last migration
migrate-down:
	@echo "?? Reverting last database migration..."
	@migrate -database "$(DB_URL)" -path migrations down

# migrate-create: creates a new migration file with the given name
# Example: make migrate-create name=add_users_table
migrate-create:
	@echo "?? Creating new migration file..."
	@migrate create -ext sql -dir migrations -seq $(name) 