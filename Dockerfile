# syntax=docker/dockerfile:1.7

# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
COPY --from=frontend-builder /app/web/dist ./web/dist
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o orbita ./cmd/server/

# Stage 3: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl && \
    adduser -D -u 10001 orbita
WORKDIR /app
COPY --from=backend-builder /app/orbita .
COPY --from=backend-builder /app/migrations ./migrations
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD curl -fsS http://localhost:8080/health || exit 1

LABEL org.opencontainers.image.title="Orbita" \
      org.opencontainers.image.description="Self-hosted multi-tenant PaaS in a single Go binary" \
      org.opencontainers.image.source="https://github.com/MUKE-coder/orbita" \
      org.opencontainers.image.licenses="MIT"

ENTRYPOINT ["./orbita"]
