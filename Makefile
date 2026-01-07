.PHONY: generate build run clean deps run-fresh docker-build docker-up docker-down docker-logs lint lint-fix migrate-create migrate-up migrate-down migrate-status

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

# Миграции (локально)
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

migrate-up:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/product_review_hub?sslmode=disable" -verbose up

migrate-down:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/product_review_hub?sslmode=disable" -verbose down

migrate-status:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/product_review_hub?sslmode=disable" version

run-fresh: generate build run
