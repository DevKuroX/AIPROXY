.PHONY: build run test lint

build:
	go build -o bin/ai_proxy ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

lint:
	go vet ./...
