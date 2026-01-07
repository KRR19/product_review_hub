.PHONY: generate build run clean deps run-fresh docker-build docker-up docker-down docker-logs lint lint-fix

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

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

run-fresh: generate build run
