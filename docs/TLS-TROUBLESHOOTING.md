# TLS / Let's Encrypt Troubleshooting

Orbita issues certificates through Traefik's ACME resolver (`letsencrypt`) using the
HTTP-01 challenge on port 80. When a certificate doesn't issue, it is almost always one
of the failure modes below — in this order of likelihood.

Use the built-in pre-flight check before debugging by hand:

```
GET /api/v1/orgs/:org/domains/verify?domain=app.example.com
```

It reports whether DNS resolves, the resolved IPs, whether Cloudflare's proxy is in
front (orange cloud), and a copy-ready fix.

---

## 1. DNS not propagated

**Symptom (Traefik log):** `unable to generate a certificate ... DNS problem: NXDOMAIN`
or `no such host`.

**Cause:** the A record doesn't exist yet or hasn't propagated.

**Fix:** create an `A` record pointing the domain at the server IP; verify with
`dig app.example.com +short` (must return the server IP), then retry. Traefik retries
automatically — no restart needed.

## 2. Cloudflare orange cloud (proxied DNS)

**Symptom:** `HTTP 404 from challenge endpoint` or `Invalid response from
http://app.example.com/.well-known/acme-challenge/...`.

**Cause:** Cloudflare's proxy answers the ACME challenge instead of Traefik.

**Fix:** in the Cloudflare DNS tab, click the orange cloud on the record to switch it to
**DNS only** (gray). Re-enable the proxy after the cert is issued if desired — but note
renewal will fail the same way; prefer "Full (strict)" mode with the origin cert, or
leave it gray. The verify endpoint detects this case (`cloudflare_proxied: true`).

## 3. Port 80/443 blocked

**Symptom:** `connection refused` / `timeout` during the challenge.

**Cause:** firewall blocks the ACME challenge.

**Fix:**
```bash
ufw allow 80/tcp && ufw allow 443/tcp && ufw reload
ss -ltnp 'sport = :80'   # Traefik must be the listener
```

## 4. Let's Encrypt rate limit

**Symptom (Traefik log):** `too many certificates already issued` / `rate limit exceeded`.

**Cause:** 5 certificates per exact domain per week (production LE).

**Fix:** wait (limits are rolling), or test with a different subdomain. Avoid
delete/re-add loops on domains — Orbita keeps the ACME storage (`acme.json`) on a
volume so re-deploys don't re-request certificates.

## 5. Traefik not seeing the route

**Symptom:** nothing ACME-related in logs at all for the domain.

**Cause:** the dynamic config file wasn't written or the file provider isn't watching.

**Fix:** confirm the route file exists on the Traefik volume
(`/etc/orbita/traefik/dynamic/<resource>--<domain-id>.json`) and that Traefik runs with
`--providers.file.directory=/etc/orbita/traefik/dynamic --providers.file.watch=true`.
Orbita rewrites route files on every successful deploy; re-deploying the app or
re-adding the domain regenerates them.

## 6. Wrong IP behind the domain

**Symptom:** challenge returns another server's content; cert never issues.

**Cause:** the A record points somewhere else (old server, load balancer).

**Fix:** compare `dig app.example.com +short` with the Orbita server IP (the verify
endpoint returns `resolved_ips`).

---

### Reading Traefik's ACME logs

```bash
cd /opt/orbita
docker compose logs traefik --tail 100 | grep -iE "acme|certificate|error"
```

### Where certificates live

`traefik_certs` volume → `/etc/traefik/acme/acme.json`. Back it up with the rest of
`/opt/orbita`; deleting it forces re-issuance for every domain (watch rate limits).
