.PHONY: dev build build-cli migrate migrate-down test lint docker-up docker-down

dev:
	air

# The control plane (server). Named orbita-server locally so it doesn't collide
# with the ./orbita CLI from build-cli; the Docker image builds its own copy and
# still calls it `orbita` inside the container.
build:
	cd web && npm run build
	go build -ldflags="-s -w" -o orbita-server ./cmd/server/

# The orbita CLI — what operators install on their own machine.
build-cli:
	go build -ldflags="-s -w" -o orbita ./cmd/orbita/

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
