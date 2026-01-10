.PHONY: generate build run clean deps run-fresh docker-build docker-up docker-down docker-logs lint lint-fix migrate-create migrate-up migrate-down migrate-status migrate-test-up migrate-test-down migrate-test-status test test-integration test-coverage test-with-docker

# Local bin
LOCAL_BIN := $(CURDIR)/bin
GOLANGCI_LINT_VERSION := v1.63.4
GOLANGCI_LINT := $(LOCAL_BIN)/golangci-lint

generate:
	cd api && oapi-codegen -config oapi-codegen.yaml openapi.yaml
	
	go mod tidy

build:
	go build -o bin/api ./cmd/api
	go build -o bin/review-watcher ./cmd/review-watcher

build-api:
	go build -o bin/api ./cmd/api

build-review-watcher:
	go build -o bin/review-watcher ./cmd/review-watcher

run:
	docker-compose up --build

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

clean:
	rm -rf bin/

deps:
	go mod download
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Linter
$(GOLANGCI_LINT):
	mkdir -p $(LOCAL_BIN)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCAL_BIN) $(GOLANGCI_LINT_VERSION)

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

lint-fix: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix ./...

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

migrate-up:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/product_review_hub?sslmode=disable" -verbose up

migrate-down:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/product_review_hub?sslmode=disable" -verbose down

migrate-status:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/product_review_hub?sslmode=disable" version

migrate-test-up:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5433/product_review_hub_test?sslmode=disable" -verbose up

migrate-test-down:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5433/product_review_hub_test?sslmode=disable" -verbose down

migrate-test-status:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5433/product_review_hub_test?sslmode=disable" version

run-fresh: generate build run

# Testing
test:
	go test -v ./...

test-integration:
	go test -v -tags=integration ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-with-docker:
	@echo "Starting test containers..."
	docker-compose up -d postgres_test
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running migrations..."
	docker-compose up migrate_test
	@echo "Running tests..."
	TEST_DB_HOST=localhost TEST_DB_PORT=5433 TEST_DB_USER=postgres TEST_DB_PASSWORD=postgres TEST_DB_NAME=product_review_hub_test go test -p 1 -v ./... || (docker-compose down && exit 1)
	@echo "Stopping containers..."
	docker-compose down
	@echo "Tests completed successfully!"

# E2E Tests only
test-e2e:
	@echo "Starting test containers..."
	docker-compose up -d postgres_test
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running migrations..."
	docker-compose up migrate_test
	@echo "Running e2e tests..."
	TEST_DB_HOST=localhost TEST_DB_PORT=5433 TEST_DB_USER=postgres TEST_DB_PASSWORD=postgres TEST_DB_NAME=product_review_hub_test go test -p 1 -v ./tests/e2e/... || (docker-compose down && exit 1)
	@echo "Stopping containers..."
	docker-compose down
	@echo "E2E tests completed successfully!"

