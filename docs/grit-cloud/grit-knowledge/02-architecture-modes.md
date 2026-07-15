# 02 — Architecture modes (the deployable shapes)

Every Grit app is scaffolded in exactly one **architecture mode**, recorded in
`grit.json`. The mode determines what containers exist and how they're built. **Orbita must
branch on the mode** — a `single` app is one container; a `triple` app is four.

## `grit.json` — the marker you detect and branch on

At the repo root of *every* Grit project:

```json
{
  "architecture": "triple",     // single | double | triple | api | mobile
  "frontend": "next",           // next | vite   (irrelevant for api mode)
  "version": "3.59.0",
  "apps": {
    "expo": false,              // React Native mobile app present?
    "desktop": false,           // Wails desktop app present?
    "docs": true                // docs site present? (--full adds this)
  }
}
```

**Detection rule:** a repo is a Grit app **iff `grit.json` exists at the root.** Read
`architecture` to pick the deploy strategy. Everything else can be derived (see
`08-grit-yaml-and-detection.md`).

## The five modes

| `architecture` | Deploy target? | Containers (app services) | Frontend |
|----------------|:--------------:|---------------------------|----------|
| `single` | ✅ VPS | **1**: one Go binary with the SPA embedded (`go:embed`) | Vite SPA, embedded |
| `double` | ✅ VPS | **2**: `api` (Go) + `web` (Next.js or Vite) | Next.js or Vite |
| `triple` | ✅ VPS | **3**: `api` (Go) + `web` (Next.js) + `admin` (Next.js) | Next.js |
| `api` | ✅ VPS | **1**: `api` (Go) only | none |
| `mobile` | ❌ not a VPS deploy | `api` (Go) + an Expo React Native app | RN (ships to app stores) |

Plus a **`--desktop`** flag that can combine with any mode: it adds a **Wails desktop app**
(`apps/desktop`) which is **not VPS-deployed** — it compiles to a native installer. And any
mode may include a **docs** site (`apps/docs`, a Next.js app) when `apps.docs` is true.

### What Orbita deploys vs. ignores

- **Deploy to the VPS:** the `api` service, and any of `web` / `admin` / `docs` that exist,
  plus the requested **addons** (postgres / redis / minio). This is what
  `docker-compose.prod.yml` describes.
- **Never deploy to the VPS:** `apps/expo` (mobile → app stores via EAS) and `apps/desktop`
  (Wails → native `.exe`/`.dmg`/`.AppImage` installers). If you see `apps.expo` or
  `apps.desktop` true, **skip those directories entirely** — they are clients of the API,
  not server workloads.

## The deployable shape of each mode

### `single` — one container
- One Go binary that serves both the JSON API **and** the compiled React SPA (via
  `//go:embed all:frontend/dist`). Go module + `main.go` live at the **repo root**;
  the SPA lives in `frontend/`.
- **Dockerfile at the repo root** (`./Dockerfile`), build context = repo root.
- Listens on **:8080**. There is **no** `docker-compose.prod.yml` and **no** separate web
  container — deliberately (a stray `apps/api/Dockerfile` would confuse auto-detectors).
- Addons still apply: point it at postgres/redis/minio via env.
- **Routing:** one domain → the single container's :8080. The SPA and the API share it
  (SPA at `/`, API at `/api/*`).

### `double` — api + web
- pnpm monorepo: `apps/api` (Go), `apps/web` (Next.js **or** Vite), `packages/shared`.
- `apps/api/Dockerfile` + `apps/web/Dockerfile` + `docker-compose.prod.yml`.
- Routing: root/`www` domain → `web:3000`; `api.` domain → `api:8080`.

### `triple` — api + web + admin (the DoD sample shape)
- pnpm monorepo: `apps/api`, `apps/web`, `apps/admin`, `packages/shared` (+ `apps/docs` if
  `--full`).
- `apps/{api,web,admin}/Dockerfile` + `docker-compose.prod.yml`.
- Routing: root → `web:3000`; `admin.` → `admin:3000`; `api.` → `api:8080`
  (+ `docs.` → `docs:3002` if present).

### `api` — Go API only
- The Go API is the whole app (no `web`/`admin`/`shared`). `apps/api/Dockerfile` +
  `docker-compose.prod.yml` (api + postgres + redis + minio only).
- Routing: root or `api.` domain → `api:8080`. Serves JSON + auto-generated API docs at
  `/docs` (Scalar).

> **Note on `api` and `single` and where the Go module lives:** `single` keeps the Go module
> at the **repo root**; the monorepo modes (`double`/`triple`/`api`) keep the Go module under
> **`apps/api/`**. Don't hard-code — detect: the Dockerfile that exists (`./Dockerfile` vs
> `./apps/api/Dockerfile`) and the location of `go.mod` tell you which. See
> `04-build-and-deploy.md`.

**Next:** [`03-project-structure.md`](./03-project-structure.md) — the real folder layout.
