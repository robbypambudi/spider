.PHONY: build test migrate sqlc tidy

BINARIES := spider-api spider-controller spider-worker spider-cli

build: $(BINARIES)

spider-api:
	go build -o bin/spider-api ./cmd/api

spider-controller:
	go build -o bin/spider-controller ./cmd/controller

spider-worker:
	go build -o bin/spider-worker ./cmd/worker

spider-cli:
	go build -o bin/spider-cli ./cmd/cli

test:
	go test -p 1 -timeout 60s ./...

migrate:
	go run ./cmd/migrate

sqlc:
	sqlc generate

tidy:
	go mod tidy
