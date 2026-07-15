# 07 — Pulse, Sentinel & GORM Studio (observability / security / DB)

## The key fact

Pulse, Sentinel, and GORM Studio are **not separate services or containers.** They are
**embedded in the Go API binary and mounted as sub-paths on the API.** There is nothing extra
to build or deploy — when the `api` container is up, they're up.

| Dashboard | What it is | URL (on the API) | Auth | Toggle |
|-----------|-----------|------------------|------|--------|
| **Pulse** | request/latency/SQL/error observability | `https://<api-domain>/pulse/ui` | basic auth | `PULSE_ENABLED` |
| **Sentinel** | WAF, rate limiting, brute-force + anomaly detection, threat feed | `https://<api-domain>/sentinel/ui` | basic auth | `SENTINEL_ENABLED` |
| **GORM Studio** | visual database browser (read/edit rows) | `https://<api-domain>/studio` | basic auth | `GORM_STUDIO_ENABLED` |
| API docs (Scalar) | interactive OpenAPI docs | `https://<api-domain>/docs` | (public/dev) | — |
| Health | liveness/readiness JSON | `https://<api-domain>/api/health` | none | — |

Each dashboard has **basic-auth credentials from env** (defaults enabled):

```
PULSE_ENABLED=true     PULSE_USERNAME=admin     PULSE_PASSWORD=<set-a-strong-one>
SENTINEL_ENABLED=true  SENTINEL_USERNAME=admin  SENTINEL_PASSWORD=<set-a-strong-one>
GORM_STUDIO_ENABLED=true GORM_STUDIO_USERNAME=admin GORM_STUDIO_PASSWORD=<set-a-strong-one>
```

## What "mount Pulse + Sentinel by default" means for Orbita (Phase 2.7)

Because they're already on the API, Orbita's work is **configuration + exposure**, not
running new workloads:

1. **Ensure they're enabled and credentialed.** Default `PULSE_ENABLED` / `SENTINEL_ENABLED`
   to `true` for Grit apps, and **generate strong `*_PASSWORD` values** if the user didn't
   supply them (never ship the `admin`/`studio`/weak defaults to production). Store them as
   encrypted secrets like any other env.
2. **Route the sub-paths.** They ride on the `api.` domain's Traefik router — no extra router
   needed. If you want them on their own hostnames (e.g. `pulse.example.com`), add a router
   that forwards to `api:8080` with the path prefix. Keep them **HTTPS-only**.
3. **Protect them.** They carry their own basic auth. For a stronger posture, additionally
   gate them behind the org's auth / an allowlist (they expose data + DB editing). At minimum:
   strong generated passwords + HTTPS.
4. **Make `grit.yaml`-toggleable.** Respect an optional `observability`/`security` toggle in
   `grit.yaml`; if absent, default on. Turning off just sets `*_ENABLED=false`.
5. **Surface the links on deploy success.** After cutover, print:
   - App: `https://<web-domain>`
   - API: `https://<api-domain>`
   - Pulse: `https://<api-domain>/pulse/ui`
   - Sentinel: `https://<api-domain>/sentinel/ui`
   - Studio: `https://<api-domain>/studio`

## GORM Studio caution

`/studio` allows editing production data directly. Enable it, but treat its password as
sensitive and consider disabling it (`GORM_STUDIO_ENABLED=false`) for hardened production —
make that a `grit.yaml` choice. (Note: Pulse/Sentinel/Studio open small embedded SQLite
stores under the container's working dir — this is why the shipped Dockerfiles `chown` `/app`
to the non-root user **before** `USER`. If you regenerate a Dockerfile, preserve that.)

**Next:** [`08-grit-yaml-and-detection.md`](./08-grit-yaml-and-detection.md).
