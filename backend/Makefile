.PHONY: build test lint

build:
	go build -o bin/api cmd/api/main.go

test:
	go test -v ./...

lint:
	golangci-lint run ./...
