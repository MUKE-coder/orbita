.PHONY: dev build build-cli migrate migrate-down test lint docker-up docker-down

dev:
	air

build:
	cd web && npm run build
	go build -ldflags="-s -w" -o orbita ./cmd/server/

# The grit CLI (grit cloud / grit deploy). Standalone binary.
build-cli:
	go build -ldflags="-s -w" -o grit ./cmd/grit/

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
