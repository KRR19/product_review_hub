.PHONY: generate build run clean deps run-fresh

generate:
	cd api && oapi-codegen -config oapi-codegen.yaml openapi.yaml

build:
	go build -o bin/api ./cmd/api

run: build
	./bin/api

clean:
	rm -rf bin/

deps:
	go mod download
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

run-fresh: generate build run
