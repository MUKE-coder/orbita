# Phase 1 — Finish & Harden Checklist

Derived from `docs/ORBITA-AUDIT.md` (§3, §6). Ordered by the Phase 1 priorities:
deploy engine → git+webhook auto-deploy → backup/restore → RBAC → TLS. Each item cites the
audit gap number. Check when done + committed + covered by a test or documented verification.

## A. Deploy engine (P1 task 2)
- [x] A1. Inject env vars (incl. decrypted secrets) into service spec on deploy (gap #1).
- [x] A2. Swarm `UpdateConfig` start-first + rollback-on-failure; container healthcheck from
      app config; failed update must not kill the running version (gap #6).
- [x] A3. Write Traefik route on deploy when the app has domains; keep ingress publish as
      fallback (gap #7).
- [x] A4. Rollback parity: set `TriggeredBy`, timestamps on rollback deployments (gap #20).
- [x] A5. Deploy edge cases return useful errors: failed build (surface build log tail),
      port mismatch, image pull failure.
- [x] A6. Ping Docker daemon at startup; fail loudly (log.Fatal in prod, warn in dev) instead
      of silent stub client (gap #17).

## B. Git + webhook auto-deploy (P1 task 3)
- [x] B1. Derive and persist `repo_url` from `repo_full_name` at app create/update (gap #22).
- [x] B2. Add `auto_deploy` (and webhook secret regeneration) to app create/update API (gap #23).
- [x] B3. Generate `webhook_secret` at git-app creation; require signature whenever secret
      exists — and it now always exists (gaps #3, #25).
- [x] B4. Fix webhook deploy: pass real org slug; record `trigger_type=webhook` (gap #24).
- [x] B5. Allow public-repo git apps without a git connection (gap #21).
- [x] B6. Webhook handler tests: valid signature → deploy; bad/missing signature → 401;
      unknown repo → ignored.

## C. Databases (P1 task 4)
- [x] C1. Attach named volume to DB service spec so data persists (gap #4).
- [x] C2. Inject `${NAME}_URL` env var into apps in the same environment on deploy (gap #5).
- [x] C3. Real backup: run engine dump in a sidecar container, gzip to local backup dir;
      record real size (gap #10).
- [x] C4. Real restore from a backup file; verify data round-trip (gap #10).
- [x] C5. Fire backup schedules via the cron scheduler (gap #10).
- [x] C6. Automated test: provision postgres → write row → backup → drop → restore → row back.

## D. RBAC + API keys (P1 tasks 5, 8)
- [x] D1. Tests: viewer cannot deploy; developer cannot manage members; cross-org access 404s;
      role hierarchy helper unit tests.
- [x] D2. Wire `orb_` API-key bearer auth into `/api/v1` org routes needed by the CLI:
      deploy, rollback, env, domains, logs, status (gap #2). Scope enforcement + tests.

## E. Secrets (P1 task 6)
- [x] E1. Unit tests: AES-256-GCM round-trip, HKDF per-org key differs per org, tampered
      ciphertext rejected.
- [x] E2. API test: secret env values never returned in plaintext (masked list response).
- [x] E3. Require `ENCRYPTION_MASTER_KEY` at startup (32-byte hex); update install.sh to
      generate 32 bytes (gap #16).

## F. TLS (P1 task 7)
- [x] F1. Document LE failure modes (orange-cloud, DNS not propagated, port 80 blocked, rate
      limit) in docs/TLS-TROUBLESHOOTING.md; surface a domain "status" with actionable error
      in the domains API (DNS lookup vs server IP already exists — extend with guidance text).

## G. Observability stubs that block real use (from audit §3 P1)
- [x] G1. Real WebSocket log streaming (docker service logs follow → WS fan-out).
- [x] G2. Real terminal (docker exec attach) — needed by existing UI.
- [x] G3. Real exec endpoint (gap #12).
- [x] G4. Real metrics for app endpoint + fix dashboard service-vs-container ID bug (gap #11).
- [x] G5. Real cron executor: run container to completion, capture logs/exit code (gap #9).

## H. Green build + release (P1 tasks 9–11)
- [x] H1. `go test ./...` green with real coverage on deploy/db/rbac/secrets paths.
- [x] H2. golangci-lint configured and clean.
- [x] H3. GHCR image published (`ghcr.io/muke-coder/orbita`) so install.sh works.
- [x] H4. Tag `v0.1.0`.

Out of scope for Phase 1 (deferred per project-description §8): GitLab/Gitea webhook parsing,
multi-node node manager, blue-green, PR previews, deploy queue activation.
