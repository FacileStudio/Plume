# Plume

Self-hosted document signing for small teams. Upload a PDF, place fields, send it to signers,
get a stamped PDF with a signature certificate and an audit trail.

Single-tenant by design: one Go binary serves the API and the SPA, one Postgres database, one
directory of uploaded PDFs. No third-party signing service sees your documents.

Live at [plume.facile.studio](https://plume.facile.studio).

## What it does

- Uploads PDFs (50 MB cap, `%PDF-` magic check) and stores them on disk under `UPLOAD_DIR`
- Places typed fields — signature, text and friends — per page, per signer, at absolute
  coordinates
- Sends signing invitations over your own SMTP server and tracks email opens with a
  tracking pixel
- Collects signatures over a tokenized public link, in parallel or in a fixed sequence
- Flattens signed fields into the PDF with `pdfcpu`, then generates a signature certificate
  and an audit trail as separate PDFs
- Verifies any PDF by its SHA-256 hash through a public, rate-limited endpoint that masks
  signer emails
- Groups documents into shared spaces with per-member roles, and files them against clients
- Fires HMAC-SHA256-signed webhooks on eleven document and signer events
- Reminds pending signers on a background ticker, at a per-user interval

## Stack

| Layer | Tech |
|---|---|
| API | Go 1.25, Chi v5, GORM, PostgreSQL 16, [tronc](https://github.com/FacileStudio/tronc) v0.6.0 |
| PDF | `go-pdf/fpdf` for generation, `pdfcpu` for overlaying and flattening |
| Client | SvelteKit 2 (Svelte 5 runes), Tailwind CSS 4, [muse](https://github.com/FacileStudio/muse), `pdfjs-dist` |
| Auth | Bearer session tokens, argon2id passwords, optional OIDC via `go-oidc/v3` |
| Deploy | Docker Compose, single distroless container behind Traefik |

## Quick start

```sh
cp .env.example .env
docker compose up -d --build
```

Compose starts `plume-db` (Postgres 16) and `plume-api`, which listens on `4000` and serves
both `/api/*` and the built SPA. Migrations run on startup.

### Local development

```sh
mise run install
cd apps/client && bun run dev          # Vite on :5173
cd apps/api    && go run .             # API on :8080 unless PORT says otherwise
```

`DATABASE_URL` is required — the API exits 1 without it.

## Configuration

| Variable | What it does |
|---|---|
| `DATABASE_URL` | Postgres connection string. Required; no default |
| `DOMAIN` | Public app URL. Drives email links, the OIDC success redirect, and CORS |
| `PORT` | HTTP listen port, `8080` by default, `4000` in Compose |
| `UPLOAD_DIR` | Where uploaded and generated PDFs live |
| `CLIENT_DIR` | Directory holding the built SPA the binary serves |
| `OIDC_ISSUER` | Set it to turn on SSO; three more `OIDC_*` variables become required |
| `SSO_ONLY` | Removes the local register and login routes entirely |

Full reference: [docs/configuration.md](docs/configuration.md).

## Structure

```
apps/
  api/       Go backend — modules/ (auth, documents, signers, fields, signing,
             verify, spaces, clients, smtp, webhooks, reminders), internal/
             (middleware, env, authcrypto, hashing, pdfutil), schemas/ (GORM)
  client/    SvelteKit SPA — adapter-static, built and served by the API binary
docs/        Architecture, configuration, development, deployment, API
scripts/     check.sh, the repository quality gate
```

## Documentation

| Doc | What's in it |
|---|---|
| [Architecture](docs/architecture.md) | Request flow, data model, signing pipeline, auth |
| [Configuration](docs/configuration.md) | Every environment variable and default |
| [Development](docs/development.md) | Local setup, tests, the quality gate |
| [Deployment](docs/deployment.md) | Docker Compose, Dokploy, Traefik routing |
| [API](docs/api.md) | HTTP endpoints and payloads |

---

Part of the [Facile Suite](https://facile.studio) — self-hosted tools for creative studios
and freelancers. One login, zero cloud dependency.
