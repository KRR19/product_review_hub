.PHONY: generate build run clean deps run-fresh docker-build docker-up docker-down docker-logs

generate:
	cd api && oapi-codegen -config oapi-codegen.yaml openapi.yaml

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

run-fresh: generate build run
