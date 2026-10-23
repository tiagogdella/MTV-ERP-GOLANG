.PHONY: build run test lint swagger

build:
		go build -o bin/server ./cmd/server

run:
		go run ./cmd/server

test:
		go test ./...

lint:
		go vet ./...
		
swagger:
		swag init -g cmd/server/main.go -o docs