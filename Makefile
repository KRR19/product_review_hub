.PHONY: generate build run clean deps run-fresh docker-build docker-up docker-down docker-logs lint lint-fix migrate-create migrate-up migrate-down migrate-status migrate-test-up migrate-test-down migrate-test-status test test-integration test-coverage test-with-docker

generate:
	cd api && oapi-codegen -config oapi-codegen.yaml openapi.yaml
	
	go mod tidy

build:
	go build -o bin/api ./cmd/api

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
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

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
	go test -v ./... || (docker-compose down && exit 1)
	@echo "Stopping containers..."
	docker-compose down
	@echo "Tests completed successfully!"

