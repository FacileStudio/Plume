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
| `internal/env` | Wraps `troncenv.Core`, adds `DOMAIN`, `UPLOAD_DIR`, `OIDC_*` |
| `internal/middleware` | `NoStore` |
| `modules/auth` | porte wiring: the `porte/kit` store over `users`, `RequireAuth`, profile and API token routes |
| `internal/authcrypto` | argon2id password hashing, 32-byte URL-safe random tokens |
| `internal/hashing` | SHA-256 over readers and files, plus a 64-hex validator |
| `internal/pdfutil` | Field flattening (fpdf overlay + pdfcpu watermark), Unicode fonts |
| `internal/documentation` | Assembles the OpenAPI document from per-module registrations |
| `modules/*` | Domain modules; each owns its router, service and types |
| `schemas/` | GORM models (the query layer), `MigrateWithIssuer` and the boot data steps in `Reconcile` |
| `migrations/` | Ordered goose SQL, embedded and applied at boot by `tronc/migrate` |

Modules are composed in `main.go`. Some register a top-level subrouter
(`documents.RegisterRoutes`), and some contribute nested routes to another module's
subrouter — `fields.DocumentRoutes`, `signers.DocumentRoutes`, `signing.DocumentRoutes` all
hang off `/api/documents`, and `reminders.SignerRoutes` hangs off `/api/signers`.

## Request lifecycle

1. `httpx.NewRouter` (tronc) applies request logging and CORS from
   `CORSAllowedOrigins`. `middleware.SecurityHeaders` is added on top.
2. `health.Mount` registers `/health` and `/ready` twice — bare and under `/api` — with a
   database ping as the readiness check.
3. `/api` routes into the module subrouters. Authenticated groups run porte/kit's
   `RequireAuth`, which checks the session cookie or `Authorization: Bearer` token through
   porte, loads the account from `users`, and puts a `kit.Identity` in the request context.
4. Anything unmatched falls through to the SPA handler.

Timeouts are set on the `http.Server`: 5s read header, 10s read, 15s write, 60s idle.
Shutdown is graceful with a 10s budget on `SIGINT`/`SIGTERM`.

Two background goroutines run for the life of the process: the reminder loop
(`reminders.Start`) and expired-session cleanup (`auth.StartSessionCleanup`). A third runs
once at boot, backfilling SHA-256 hashes for documents stored before hashing existed.

## Data model

The schema is the ordered SQL under `apps/api/migrations/`, applied at boot by `tronc/migrate`
behind a Postgres advisory lock. `00001_baseline.sql` is the production schema GORM
`AutoMigrate` had built, as `pg_dump` printed it on 2026-11-13; `apps/api/schema.sql` is the
normalized dump of the whole set. The GORM structs no longer create anything:
`TestGORMModelsMatchTheMigratedSchema` fails when a struct names a column the migrations do
not create, or a type its field cannot hold. The porte tables, `directory_deletion_state` and
tronc/mcp's `mcp_actions` are in the baseline too; `TestThePorteTablesMatchPortePG` and
`TestTheMCPSchemaIsInTheMigrations` fail when a library bump changes their shape.

After the migrations, `schemas.Reconcile` runs the idempotent data steps the old boot code
ran, in the same order: a `facile_id` for any space without one, the avatar column backfill,
and `AdoptPorte` (the local identity re-key, the pre-porte session carry-over, the password
and OIDC subject adoption, the Registre picture backfill).

The domain tables, all with `int64` primary keys:

| Table | Key columns | Relationships |
|---|---|---|
| `users` | `email` unique, `password_hash` (legacy, unread), `locale` (default `en`), `reminder_interval_days` (default 3), `avatar_source`, `oidc_picture_url`, `oidc_access_token`, `oidc_refresh_token`, `profile_synced_at` | owns everything else |
| `porte_sessions`, `porte_identities` | porte's sessions and sign-in identities, keyed on `users.id` | cascade on user delete |
| `spaces` | `name`, `description`, `owner_id` | groups documents |
| `space_members` | `space_id`, `user_id`, `role` (default `member`) | unique on `(space_id, user_id)`, cascades on delete |
| `documents` | `status` (default `draft`), `storage_path`, `original_hash`, `signed_hash`, `owner_id`, `space_id`, `client_id`, `sequential` | has many signers and fields |
| `signers` | `document_id`, `email`, `role` (default `signer`), `status` (default `pending`), `token` (sealed when a key is set), `token_hash` (indexed), `order_num`, `signed_at`, `viewed_at`, `email_opened_at`, `ip_address`, `user_agent`, `last_reminded_at` | belongs to a document |
| `fields` | `document_id`, `signer_id`, `field_type`, `page`, `x`, `y`, `width`, `height`, `required`, `label`, `value` | belongs to a document and a signer |
| `clients` | `name`, `email`, `company`, `phone`, `notes`, `owner_id` | documents file against one |
| `webhooks` | `url`, `secret` (sealed when a key is set), `enabled`, `last_sent_at` | per owner |
| `smtp_configs` | `owner_id` unique, `host`, `port`, `username`, `password` (sealed when a key is set), `from_email`, `from_name` | one per owner |

`space_members` is unique on `(space_id, user_id)`, `spaces.facile_id` and
`spaces.registre_space_id` are unique, and `documents` is unique on `(owner_id, external_ref)`
when `external_ref` is set.

Document status moves `draft` → `pending` → `completed` or `declined`. Only a `draft` or
`declined` document can be deleted (`DELETE /api/documents/{id}`, `plume_delete_document`);
a `pending` or `completed` one answers 400 `invalid_argument`. Erasing an account or a space
is a separate path and still removes everything.

### Secrets at rest

Three values are stored for later use and are protected in the database:

| Value | Column | What is stored |
|---|---|---|
| Signing link token | `signers.token_hash` | SHA-256 hex of the token, always |
| Signing link token | `signers.token` | the token sealed with `PORTE_TOKEN_KEY`; the token itself when no key is set |
| SMTP password | `smtp_configs.password` | sealed with `PORTE_TOKEN_KEY`; as typed when no key is set |
| Webhook secret | `webhooks.secret` | sealed with `PORTE_TOKEN_KEY`; as typed when no key is set |

Every lookup by signing token goes through `secrets.FindSigner`, which matches
`token_hash = sha256(input)`. The hash is written with or without a key. The token column
still holds a usable token because reminders, the signing email and the owner's copy-link
need it back, so those paths open it with the key at the point of use. API, MCP and webhook
responses keep returning the usable link.

Sealing is porte's token sealer (`porte/pg`): AES-256-GCM under a key derived from
`PORTE_TOKEN_KEY` with HKDF, stored as `pv1:` and the base64url of nonce and ciphertext. The
same key already seals the OIDC tokens in `porte_identities`. A comma-separated list rotates
the key: the first seals, every one opens. A sealed value from the database is never accepted
as a signing token, whatever the input: the lookup falls back to the plaintext `token` column
only for rows not converted yet, and only for an input that has no `pv1:` prefix and no colon.

At boot, after the migrations, a background pass converts rows written by older releases. It
fills `token_hash` for every signer that lacks one and, when a key is set, seals `token`,
`smtp_configs.password` and `webhooks.secret`. Each row is rewritten by an update conditioned
on the value the pass read, so two replicas running it at once change a row once, and a
second run changes nothing. It logs counts only. With no key (self-hosting) the pass only
hashes, and the three values stay in clear.

Losing or changing `PORTE_TOKEN_KEY` without keeping the old key in the list makes the sealed
values unreadable. A send then fails with a clear error instead of using the ciphertext as a
password, and webhook deliveries fail the same way. Recover by entering each SMTP password and
webhook secret again in Settings, and by re-sending the pending signing links (a signer's
`token_hash` still works for a link already in an inbox, but Plume can no longer rebuild that
link for a reminder or the copy-link button).

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
   Signing and declining are atomic. One transaction locks the document row, claims the
   signer with `UPDATE signers ... WHERE id = ? AND status = 'pending'`, writes the field
   values and sets the document status. A second request with the same token claims
   nothing, gets "already signed or declined", and writes nothing. Webhooks, emails, the
   Nuage upload and cache invalidation run only after the commit.
5. **Flatten.** Stamping happens lazily, when someone downloads the file.
   `pdfutil.FlattenFields` builds a transparent overlay PDF with `fpdf` at 1:1 page
   dimensions and applies it to the original as a pdfcpu watermark. pdfcpu does the merge
   because it reads arbitrary real-world PDFs, which fpdf's page importer does not. The
   stamped file is cached next to the original, in one of two places. While the document is not
   `completed`, a throwaway rendering goes to `<name>_pending.pdf`. It is regenerated whenever
   the original is newer, removed on every signature, and never hashed. Once the document is
   `completed`, the signed file `<name>_signed.pdf` is created exactly once: the first render
   goes to a unique temporary file and is hard-linked into place, which fails if the target
   exists, so concurrent downloads and the Nuage publish job, on any replica sharing the
   volume, end up with the same bytes. After that it is frozen. It is served whatever the file
   times say, never re-rendered and never deleted by a signature. Rendering is not
   reproducible, and a new render would change the hash of every copy already handed out.
   `signed_hash` is the SHA-256 of that file, written once while the column is empty and never
   overwritten. If the file is missing while `signed_hash` is set, the download renders a new
   file, keeps the stored hash and logs an error with the document id. A `_signed.pdf` that an
   older release left beside a still-pending document is removed once, inside the transaction
   that completes it, while `signed_hash` is empty.
6. **Attest.** `GET /api/documents/{id}/certificate` and
   `GET /api/documents/{id}/audit-trail` generate two more PDFs on first request and cache
   them next to the document. Editing a document invalidates both.

## The signer audit trail

When a signer signs or declines, `modules/signers` stamps the signer row with `signed_at`
(the time of the answer), `ip_address` and `user_agent`. The certificate and the audit-trail
PDFs print all three. Since v0.13.4, `ip_address` is the host of `RemoteAddr` after tronc's
real-IP middleware, which reads the forwarded address only from peers in `TRUSTED_PROXIES`
(`private,cloudflare` in production).

### Rows that hold a proxy address

Rows written before v0.13.4 do not hold the signer's address. The routes read `X-Real-IP` by
hand, and in production that header carried the address of a proxy hop inside the server.
These rows are every `signers` row with a non-empty `ip_address` and a `signed_at` before the
v0.13.4 deploy at 2026-11-12 10:57 UTC:

```sql
SELECT id FROM signers WHERE ip_address <> '' AND signed_at < '2026-11-12 10:57:38+00';
```

A read-only count on 2026-11-13 found 23 such rows, all with status `signed`. Their
`created_at` runs from 2026-05-14 17:23 to 2026-07-03 15:30 UTC and their `signed_at` from
2026-05-14 17:54 to 2026-07-21 15:03 UTC. Every one holds one of four addresses in
`10.0.1.0/24`, the private subnet of the Docker overlay network that connects the reverse
proxy to Plume.

These rows, and the certificate and audit-trail PDFs that print them, stay unchanged. The
audit trail is evidence: a value rewritten after the fact would make every trail less
trustworthy, and the signer's real address was never recorded, so there is nothing true to
write. For these rows the IP shows only that the request passed through Plume's reverse
proxy. It identifies no signer, device or network. The rest of the record is unaffected:
`signed_at`, the user agent, the view and email-open times and the document hashes are what
they always were.

## Verification

`POST /api/verify` and `GET /api/verify/{hash}` are public and unauthenticated. They take a
64-character lowercase hex SHA-256 and look for a document whose `original_hash` or
`signed_hash` matches, reporting which of the two variants hit. The response carries the
document name, status and timestamps plus the signer list, with every signer email masked.
The route group runs behind tronc's `middleware.Ratelimit` (per visitor address: a burst of 10, then
one every 3 seconds, so at most 30 in any minute; a refusal is 429 `rate_limited` with a computed
`Retry-After`) and `NoStore`, which sets
`Cache-Control: no-store`, `Referrer-Policy: no-referrer` and `X-Content-Type-Options`.

## Authentication

Sign-in goes through Registre only. porte mounts no register, login or password route while
an issuer is configured, and `OIDC_ISSUER` is required in production. A browser session is a
porte cookie session; API clients (Ardoise) send an API token, a labelled porte session minted
from Settings > API, as `Authorization: Bearer`. Signing links (`/api/sign/{token}`) and
`/api/verify` need no session.

`GET /api/auth/config` tells the client what is available: `oidc_enabled`, and when OIDC is
on, the issuer and redirect URL.

Registre erases people on request. `POST /api/registre/deletions` (a signed `user.deleted`
push, no session) and a poll of Registre's deletion list both call `spaces.Deletions`, which
deletes the account, its sessions, API tokens, SMTP settings, webhooks, unused contacts and
unsent drafts, and keeps every sent or signed document byte-identical. See
[erasure.md](erasure.md) for the field-by-field list.

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
- **Webhooks.** Plume does not speak the Nook `pool`/`enveloppe` protocol. It dispatches its
  own JSON webhooks to per-owner URLs, signed with HMAC-SHA256 over the body and sent as
  `x-plume-signature-256: sha256=<hex>` with `User-Agent: Plume-Webhook/1.0`. Eleven event
  types fire:
  `document.created`, `document.sent`, `document.completed`, `document.declined`,
  `document.deleted`, `signer.added`, `signer.email_opened`, `signer.viewed`,
  `signer.signed`, `signer.declined` and `signer.reminded`. Payload shapes are in
  [api.md](api.md).
- **Porte.** OIDC federates to Authentik at `porte.facile.studio` like the rest of the
  suite, with its own application slug.
