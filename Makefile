.PHONY: dev build migrate migrate-down test lint docker-up docker-down

dev:
	air

build:
	cd web && npm run build
	go build -ldflags="-s -w" -o orbita ./cmd/server/

migrate:
	go run ./cmd/migrate/main.go up

migrate-down:
	go run ./cmd/migrate/main.go down

test:
	go test ./... -v -race

lint:
	golangci-lint run

docker-up:
	docker compose -f docker/docker-compose.dev.yml up -d

docker-down:
	docker compose -f docker/docker-compose.dev.yml down
