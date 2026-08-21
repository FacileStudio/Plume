# Plume — Architecture

How a request reaches Plume, what the database holds, how a PDF becomes a signed PDF, and how
authentication works.

## Runtime topology

```
Internet ──▶ Traefik ──▶ Go binary (:4000) ──┬──▶ /health, /ready      liveness + readiness
                                              ├──▶ /api/health, /ready  same, under /api
                                              ├──▶ /api/*               module handlers
                                              └──▶ /*                   SPA catch-all
                                                              │
                                          ┌───────────────────┼───────────────────┐
                                    Postgres 16        /data/uploads        SMTP server
                                    (GORM)             (PDFs, avatars)      (per user)
                                                              │
                                              Journal (log shipping, optional)
                                              Authentik (OIDC, optional)
```

One container, one router, one hostname. The client is built with
`@sveltejs/adapter-static` and copied into the API image at `/client`; `tronc/spa` serves it
as the catch-all. The browser calls `/api/*` on the same origin, so there is no proxy hop
and CORS is only relevant when you run the Vite dev server on `:5173`.

## Components

| Package | Responsibility |
|---|---|
| `main.go` | Config load, logger, DB open, migrations, service wiring, router, shutdown |
| `internal/env` | Wraps `troncenv.Core`, adds `DOMAIN`, `UPLOAD_DIR`, `SSO_ONLY`, `OIDC_*` |
| `internal/middleware` | `RequireAuth`, `SecurityHeaders`, `NoStore`, in-process rate limiter |
| `internal/authcrypto` | argon2id password hashing, 32-byte URL-safe random tokens |
| `internal/hashing` | SHA-256 over readers and files, plus a 64-hex validator |
| `internal/pdfutil` | Field flattening (fpdf overlay + pdfcpu watermark), Unicode fonts |
| `internal/documentation` | Assembles the OpenAPI document from per-module registrations |
| `modules/*` | Domain modules; each owns its router, service and types |
| `schemas/` | GORM models and `Migrate` |

Modules are composed in `main.go`. Some register a top-level subrouter
(`documents.RegisterRoutes`), and some contribute nested routes to another module's
subrouter — `fields.DocumentRoutes`, `signers.DocumentRoutes`, `signing.DocumentRoutes` all
hang off `/api/documents`, and `reminders.SignerRoutes` hangs off `/api/signers`.

## Request lifecycle

1. `httpx.NewRouter` (tronc) applies request logging and CORS from
   `CORSAllowedOrigins`. `middleware.SecurityHeaders` is added on top.
2. `health.Mount` registers `/health` and `/ready` twice — bare and under `/api` — with a
   database ping as the readiness check.
3. `/api` routes into the module subrouters. Authenticated groups run
   `middleware.RequireAuth`, which reads the `Authorization: Bearer` token, looks the
   session up, and puts the identity in the request context.
4. Anything unmatched falls through to the SPA handler.

Timeouts are set on the `http.Server`: 5s read header, 10s read, 15s write, 60s idle.
Shutdown is graceful with a 10s budget on `SIGINT`/`SIGTERM`.

Two background goroutines run for the life of the process: the reminder loop
(`reminders.Start`) and expired-session cleanup (`auth.StartSessionCleanup`). A third runs
once at boot, backfilling SHA-256 hashes for documents stored before hashing existed.

## Data model

Ten tables, all `int64` primary keys, migrated by GORM `AutoMigrate` on every start.

| Table | Key columns | Relationships |
|---|---|---|
| `users` | `email` unique, `password_hash`, `reminder_interval_days` (default 3), `avatar_source`, `oidc_picture_url`, `oidc_access_token`, `oidc_refresh_token`, `profile_synced_at` | owns everything else |
| `sessions` | `token` primary key, `user_id`, `expires_at` | belongs to a user |
| `spaces` | `name`, `description`, `owner_id` | groups documents |
| `space_members` | `space_id`, `user_id`, `role` (default `member`) | unique on `(space_id, user_id)`, cascades on delete |
| `documents` | `status` (default `draft`), `storage_path`, `original_hash`, `signed_hash`, `owner_id`, `space_id`, `client_id`, `sequential` | has many signers and fields |
| `signers` | `document_id`, `email`, `role` (default `signer`), `status` (default `pending`), `token`, `order_num`, `signed_at`, `viewed_at`, `email_opened_at`, `ip_address`, `user_agent`, `last_reminded_at` | belongs to a document |
| `fields` | `document_id`, `signer_id`, `field_type`, `page`, `x`, `y`, `width`, `height`, `required`, `label`, `value` | belongs to a document and a signer |
| `clients` | `name`, `email`, `company`, `phone`, `notes`, `owner_id` | documents file against one |
| `webhooks` | `url`, `secret`, `enabled`, `last_sent_at` | per owner |
| `smtp_configs` | `owner_id` unique, `host`, `port`, `username`, `password`, `from_email`, `from_name` | one per owner |

`Migrate` also drops the legacy non-unique `idx_signers_token` index and creates the unique
`idx_space_members_space_user` index.

Document status moves `draft` → `pending` → `completed` or `declined`.

## The signing pipeline

1. **Upload.** `POST /api/documents` takes a multipart form. The handler caps the body at
   50 MB, reads the first five bytes and rejects anything that is not `%PDF-`, writes the
   file to `UPLOAD_DIR/<owner_id>/<doc_id>_<filename>` while hashing it, and stores the
   SHA-256 as `original_hash`.
2. **Prepare.** Fields are created against the document with a page number and absolute
   coordinates, each assigned to one signer. Signers are added with an `order_num`.
3. **Send.** `POST /api/documents/{id}/send` accepts only a `draft` with at least one signer.
   It mints a token for every signer, moves the document to `pending`, and emails the
   signing link through the owner's SMTP config. Only signers whose role is `signer` or
   `approver` are invited; on a `sequential` document, only those sharing the lowest
   `order_num`.
4. **Sign.** The signer opens `/share/<token>`, which reads
   `GET /api/sign/{token}` and `GET /api/sign/{token}/file`. Submitting posts field values
   and a signature image data URL to `POST /api/sign/{token}`, recording IP and user agent.
5. **Flatten.** Stamping happens lazily, when someone downloads the file.
   `pdfutil.FlattenFields` builds a transparent overlay PDF with `fpdf` at 1:1 page
   dimensions and applies it to the original as a pdfcpu watermark. pdfcpu does the merge
   because it reads arbitrary real-world PDFs, which fpdf's page importer does not. The
   stamped file is cached next to the original and regenerated whenever the original is
   newer. `signed_hash` is only persisted once the document reaches `completed`, so a
   partially signed intermediate never becomes the authoritative fingerprint.
6. **Attest.** `GET /api/documents/{id}/certificate` and
   `GET /api/documents/{id}/audit-trail` generate two more PDFs on first request and cache
   them next to the document. Editing a document invalidates both.

## Verification

`POST /api/verify` and `GET /api/verify/{hash}` are public and unauthenticated. They take a
64-character lowercase hex SHA-256 and look for a document whose `original_hash` or
`signed_hash` matches, reporting which of the two variants hit. The response carries the
document name, status and timestamps plus the signer list, with every signer email masked.
The route group runs behind a 30-requests-per-minute rate limiter and `NoStore`, which sets
`Cache-Control: no-store`, `Referrer-Policy: no-referrer` and `X-Content-Type-Options`.

## Authentication

Local auth is email and password. Passwords are argon2id (64 MiB, 3 iterations, parallelism
2, 16-byte salt, 32-byte key) and compared in constant time. A successful login mints a
32-byte URL-safe random token stored in `sessions`; the client keeps it in `localStorage`
and sends it as `Authorization: Bearer`. Register and login are rate-limited to 10 requests
per minute. Setting `SSO_ONLY=true` does not merely reject local logins — the routes are
never registered.

`GET /api/auth/config` tells the client what is available: `sso_only`, `oidc_enabled`, and
when OIDC is on, the issuer and redirect URL.

### The OIDC flow, and why Plume's differs

OIDC turns on when `OIDC_ISSUER` is set; the other three `OIDC_*` variables then become
required and startup fails without them. Discovery runs at boot through `go-oidc/v3`, with
scopes `openid email profile offline_access`.

```
browser ──▶ GET /api/auth/oidc            sets oidc_state cookie, redirects to Authentik
Authentik ─▶ GET /api/auth/oidc/callback  state compared in constant time, code exchanged,
                                          id_token verified, email + email_verified checked,
                                          user upserted, one-time code minted (60s TTL)
browser ──▶ OIDC_SUCCESS_URL?code=<code>  SPA reads the query parameter
SPA ──────▶ POST /api/auth/oidc/exchange  {"code": ...} → {"user_id", "token"}
```

**This is the suite exception.** Every other Facile app hands the session token back in the
URL fragment. Plume instead stores a pending code in an in-process `sync.Map`, redirects with
`?code=` in the query string, and requires a second POST to trade it for the session token.
The token therefore never appears in a URL at all. The cost is that pending codes live in
process memory, so the callback and the exchange must hit the same instance — Plume does not
horizontally scale without sticky sessions. A ticker sweeps expired codes every five minutes.

`POST /api/auth/sync-profile` re-reads the userinfo endpoint for the logged-in user;
`profile_synced_at` rate-limits it. Avatars resolved from the `picture` claim are downloaded
into `UPLOAD_DIR/avatars` and served from `GET /api/files/*`.

## Cross-app integration

- **Journal.** When both `JOURNAL_URL` and `JOURNAL_TOKEN` are set, the tronc logger is
  wrapped in `journal.NewHandler` and every structured log line ships to Journal. Setting
  only one of the two ships nothing.
- **Webhooks.** Plume does not speak the Antenne `pool`/`enveloppe` protocol. It dispatches its
  own JSON webhooks to per-owner URLs, signed with HMAC-SHA256 over the body and sent as
  `x-plume-signature-256: sha256=<hex>` with `User-Agent: Plume-Webhook/1.0`. Eleven event
  types fire:
  `document.created`, `document.sent`, `document.completed`, `document.declined`,
  `document.deleted`, `signer.added`, `signer.email_opened`, `signer.viewed`,
  `signer.signed`, `signer.declined` and `signer.reminded`. Payload shapes are in
  [api.md](api.md).
- **Porte.** OIDC federates to Authentik at `porte.facile.studio` like the rest of the
  suite, with its own application slug.
