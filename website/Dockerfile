# Orbita docs website — production image for Dokploy (or any Docker host).
#
# Build context is the `website/` directory. In Dokploy, set:
#   Build type: Dockerfile
#   Docker file: Dockerfile
#   Build context / directory: website
# Then point a domain at the app; Dokploy's Traefik terminates TLS. The app
# listens on port 80.

# ---- Stage 1: build the static site ----
FROM node:20-alpine AS builder
WORKDIR /app

# Install deps from the lockfile for a reproducible build.
COPY package.json package-lock.json ./
RUN npm ci

# Build the VitePress site → .vitepress/dist
COPY . .
RUN npm run build

# ---- Stage 2: serve with nginx ----
FROM nginx:1.27-alpine

# VitePress clean-URL routing + asset caching.
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Static output only — no Node runtime in the final image.
COPY --from=builder /app/.vitepress/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost/ >/dev/null 2>&1 || exit 1

CMD ["nginx", "-g", "daemon off;"]
