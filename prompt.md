# Claude Code Prompt — Orbita PaaS

Paste this entire prompt into Claude Code to begin building Orbita.

---

## Prompt

You are going to help me build **Orbita** — a production-grade, self-hosted PaaS (Platform as a Service) written in Go with an embedded React SPA frontend.

I have provided two reference documents you must read before writing any code:
- `project-description.md` — the full product specification, tech stack, data models, API structure, and feature definitions
- `project-phases.md` — the detailed phased build plan with every task you must complete

Read both documents carefully before starting. Every decision about naming, structure, tooling, and architecture must follow what is described in those documents.

---

## Your Tech Stack (do not deviate from this)

**Backend**
- Go 1.22+
- Gin (HTTP framework)
- GORM (ORM) with PostgreSQL driver
- PostgreSQL 15 (database)
- Redis 7 (cache, queue, sessions)
- JWT (golang-jwt/jwt v5) — access + refresh tokens
- Resend (email — resend-go/v2)
- Docker SDK (docker/client) — container management
- gorilla/websocket — real-time logs and terminal
- robfig/cron v3 — cron job scheduler
- golang-migrate/migrate — DB migrations
- zerolog — structured logging
- godotenv — env config

**Frontend** (inside `web/` directory)
- React 18 + TypeScript
- Vite (build tool)
- Tailwind CSS v3
- shadcn/ui (component library)
- Zustand (state management)
- TanStack Query / React Query (data fetching)
- React Router DOM v6
- React Hook Form + Zod (forms + validation)
- Lucide React (icons)
- Recharts (charts/metrics)
- xterm.js (SSH terminal)

**Build output**: Single Go binary with embedded React SPA (`//go:embed all:web/dist`)

---

## Rules You Must Follow

### Architecture rules
1. **Layered architecture**: handlers → service → repository → database. Business logic lives in service layer only. Handlers only validate input and call services. Repositories only do DB queries.
2. **Org scoping is mandatory**: Every query that touches user data must include `organization_id` in the WHERE clause. Use a GORM scope `OrgScope(orgID)` applied at the repository layer. This is non-negotiable for security.
3. **No ORM magic for security queries**: Scope queries explicitly — do not rely on GORM associations to implicitly scope by org.
4. **Secrets encrypted at rest**: Any value marked as a secret in env vars must be encrypted with AES-256-GCM using the org-derived key before storing. Never store plaintext secrets in the database.
5. **Context everywhere**: Every function that does I/O must accept `context.Context` as its first parameter.
6. **Error wrapping**: Use `fmt.Errorf("doThing: %w", err)` everywhere. Return errors; do not panic except in `main.go` startup.

### Go code rules
7. One file per handler group (e.g., `auth_handler.go`, `app_handler.go`). No god files.
8. All GORM models must have: `ID uuid.UUID` (primary key, `gorm:"type:uuid;default:uuid_generate_v4()"`), `CreatedAt`, `UpdatedAt`, `DeletedAt gorm.DeletedAt` (soft delete).
9. Use `uuid.UUID` everywhere for IDs — never integer IDs.
10. Validate all incoming JSON/form data using Gin's `ShouldBindJSON` with struct tags (`binding:"required,email"` etc.).
11. Standardize API responses: success `{"data": ...}`, error `{"error": {"code": "...", "message": "..."}}`.

### Frontend rules
12. All API calls go through a typed API client in `web/src/api/`. No raw fetch calls in components.
13. All forms use React Hook Form + Zod schema validation. No manual state for forms.
14. Loading states must be handled. Every button that triggers an API call must show a spinner and be disabled while loading.
15. Error states must be handled. API errors must display a user-friendly message (toast notification or inline error).
16. Use React Query for all server state. No useState for data that comes from the API.
17. The sidebar must show: org switcher, project list, and nav items. It must be responsive (collapsible on mobile).
18. shadcn/ui components must be used for all UI elements (buttons, inputs, modals, dropdowns, tables, etc.).

### Security rules
19. JWT access tokens: 15 minute TTL. Refresh tokens: 30 day TTL, stored in httpOnly SameSite=Strict cookie.
20. Rate limit all auth endpoints: 5 attempts per 15 minutes per IP using Redis sliding window counter.
21. CORS: only allow the frontend origin. In development: `http://localhost:5173`. In production: the configured `APP_BASE_URL`.
22. All WebSocket connections must authenticate via token passed as query parameter (`?token=`). Validate on handshake, not after.
23. Passwords must be hashed with bcrypt cost 12 minimum.
24. Log all sensitive actions in the `audit_logs` table (member changes, deploys, rollbacks, terminal access, backup restores).

---

## How to Work

Work through `project-phases.md` one phase at a time. Complete all tasks in a phase before moving to the next. For each phase:

1. Read the tasks list for that phase
2. Create the required files
3. Ensure the code compiles before moving on
4. If a task depends on Docker or external services during development, stub it with a commented `// TODO: real impl` that returns a sensible mock value so the API is testable without Docker
5. Write migrations before writing models — they must be consistent

When you finish a phase, tell me:
- What was built
- What endpoints are now available
- What I can test
- What the next phase will build

---

## Development Environment Setup

Before starting Phase 0, create the following:

**`.env.example`** (template for developers):
```
# Server
SERVER_PORT=8080
APP_BASE_URL=http://localhost:8080
IS_PRODUCTION=false

# Database
DATABASE_URL=postgres://orbita:orbita@localhost:5432/orbita?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# Auth
JWT_SECRET=change-me-to-a-random-64-char-string
JWT_REFRESH_SECRET=change-me-to-a-different-random-64-char-string
ENCRYPTION_MASTER_KEY=change-me-to-a-random-32-byte-hex

# Email (Resend)
RESEND_API_KEY=re_xxxxxxxx
RESEND_FROM_EMAIL=orbita@yourdomain.com

# Docker
DOCKER_SOCKET=/var/run/docker.sock

# Traefik
TRAEFIK_CONFIG_DIR=/etc/orbita/traefik

# Super Admin (first run)
SUPER_ADMIN_EMAIL=admin@example.com
```

**`docker/docker-compose.dev.yml`** for local development:
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: orbita
      POSTGRES_USER: orbita
      POSTGRES_PASSWORD: orbita
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

**`Makefile`**:
```makefile
.PHONY: dev build migrate test lint docker-up docker-down

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
```

---

## Start Now

Begin with **Phase 0: Project Scaffold & Tooling** from `project-phases.md`. Complete every task in Phase 0 before asking me whether to proceed to Phase 1.

Do not write placeholder code for future phases. Focus entirely on the current phase. Each line of code you write should compile and serve its purpose in the current phase.

When you are done with Phase 0, I will review and tell you to proceed to Phase 1.

Let's build Orbita.
