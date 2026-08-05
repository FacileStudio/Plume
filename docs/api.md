# Plume — API

Every HTTP route the Go binary registers, grouped by module, generated from the routers in
`apps/api/modules/`.

All application routes live under `/api`. Responses are JSON via `httpjson.WriteJSON`, and
errors share the tronc error shape. Authenticated routes want
`Authorization: Bearer <token>`; the token comes from login, register, or the OIDC exchange.

Plume also serves its own reference at `GET /docs` — a Scalar UI over the OpenAPI 3.1 document
at `GET /docs/openapi.json` — mounted on the root router through `tronc/apiref`, beside `/api`
rather than behind it. The document is generated from the route registry in
`apps/api/modules/*/documentation.go`, and `TestEveryRouteIsDocumented` fails the build if a
registered route is missing from it.

## Health

`/health` and `/ready` are mounted twice, bare and under `/api`, all unauthenticated.
`/ready` pings the database; `/health` does not.

## Auth

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/api/auth/config` | no | `{ sso_only, oidc_enabled }`, plus `oidc_issuer` and `oidc_redirect_url` when OIDC is on |
| POST | `/api/auth/register` | no | `{ email, password }` → `201 { user_id, token }`. Not registered when `SSO_ONLY` |
| POST | `/api/auth/login` | no | `{ email, password }` → `{ user_id, token }`. Not registered when `SSO_ONLY` |
| GET | `/api/auth/me` | yes | Profile: `id`, `email`, `name`, `avatar_url`, `avatar_source`, `reminder_interval_days`, `created_at` |
| PUT | `/api/auth/me` | yes | `{ name, email, reminder_interval_days? }` |
| PUT | `/api/auth/password` | yes | `{ current_password, new_password }` → `{ status: "ok" }` |
| GET | `/api/auth/oidc` | no | Sets the `oidc_state` cookie and redirects to the provider |
| GET | `/api/auth/oidc/callback` | no | Provider redirect target; ends by redirecting to `OIDC_SUCCESS_URL?code=…` |
| POST | `/api/auth/oidc/exchange` | no | `{ code }` → `{ user_id, token }`. Code is single-use and expires after 60s |
| POST | `/api/auth/sync-profile` | yes | Re-reads the provider's userinfo for the current user |

`/register` and `/login` share a 10-requests-per-minute rate limiter. The four OIDC routes
only exist when `OIDC_ISSUER` is set.

## Files

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/api/files/*` | no | Static server over `UPLOAD_DIR`, `Cache-Control: public, max-age=86400, immutable`. Serves avatars |

## Documents

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/api/documents` | yes | Multipart with `name`, `file`, optional `space_id`. 50 MB cap, `%PDF-` header required → `201` |
| GET | `/api/documents` | yes | Query filters `status`, `space_id`, `client_id` |
| GET | `/api/documents/stats` | yes | `{ total, pending, completed }` |
| GET | `/api/documents/{id}` | yes | One document |
| PUT | `/api/documents/{id}` | yes | `{ name, file_name, sequential?, client_id? }` |
| DELETE | `/api/documents/{id}` | yes | Fires `document.deleted` |
| GET | `/api/documents/{id}/file` | yes | The PDF. A `draft` serves the untouched upload; any other status serves the stamped version with the field values collected so far |
| POST | `/api/documents/{id}/send` | yes | Draft only, at least one signer. Moves to `pending` and emails invitations |

`DocumentResponse` carries `id`, `name`, `status`, `file_name`, `owner_id`, `sequential`,
`client_id`, `original_hash`, `signed_hash`, `created_at`, `updated_at`.

## Fields

Nested under `/api/documents`.

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/api/documents/{docId}/fields` | yes | All fields on a document |
| POST | `/api/documents/{docId}/fields` | yes | `{ signer_id, field_type, page, x, y, width, height, required, label }` |
| PUT | `/api/documents/{docId}/fields/{fieldId}` | yes | Same shape without `signer_id` |
| DELETE | `/api/documents/{docId}/fields/{fieldId}` | yes | |

Coordinates are absolute, in the same units the flattener draws with.

## Signers

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/api/documents/{docId}/signers` | yes | Roster for a document |
| POST | `/api/documents/{docId}/signers` | yes | `{ name, email, role, order }` |
| DELETE | `/api/signers/{id}` | yes | |
| POST | `/api/signers/{id}/remind` | yes | Manual reminder → `{ status, reminded_at }` |

## Signing (public, token-scoped)

These carry no session. The signer's token from the invitation email is the credential.

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/api/sign/{token}` | token | `{ document, signer, fields, completed_fields }` |
| GET | `/api/sign/{token}/status` | token | Progress roster — name, role, status, order, `signed_at`, `is_you`. No emails, IPs or user agents |
| GET | `/api/sign/{token}/opened.gif` | token | Tracking pixel; stamps `email_opened_at` and fires `signer.email_opened` |
| GET | `/api/sign/{token}/file` | token | The PDF to sign, `Content-Type: application/pdf` |
| POST | `/api/sign/{token}` | token | `{ fields: [{ field_id, value }] }` → `{ status: "signed" }`. Records IP (`X-Real-IP` when present) and user agent |
| POST | `/api/sign/{token}/decline` | token | Moves the signer and document to `declined` |

## Certificates

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/api/documents/{docId}/certificate` | yes | Signature certificate PDF, generated on first request and cached |
| GET | `/api/documents/{docId}/audit-trail` | yes | Audit trail PDF, same caching |

Editing a document invalidates both.

## Verification (public)

Behind a 30-requests-per-minute limiter and `Cache-Control: no-store`.

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/api/verify` | no | Multipart with a `file` part. The file is hashed in the request stream and never stored |
| GET | `/api/verify/{hash}` | no | 64-character hex SHA-256, lowercased before lookup |

Both return the same shape:

```json
{
  "match": true,
  "hash": "…",
  "variant": "signed",
  "document": { "name": "…", "file_name": "…", "status": "completed",
                "created_at": "…", "completed_at": "…" },
  "signers": [{ "name": "…", "email": "j***@e***.com", "status": "signed",
                "signed_at": "…" }]
}
```

`variant` is `original` or `signed`, saying which hash matched. Signer emails are masked. A
miss returns `{ "match": false, "hash": "…" }`.

## Spaces

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/api/spaces` | yes | `{ name, description }` |
| GET | `/api/spaces` | yes | Every space you are a member of, newest first; each carries your `role` |
| GET | `/api/spaces/{spaceId}` | yes | |
| PUT | `/api/spaces/{spaceId}` | yes | `{ name, description }` |
| DELETE | `/api/spaces/{spaceId}` | yes | |
| POST | `/api/spaces/{spaceId}/leave` | yes | |
| GET | `/api/spaces/{spaceId}/members` | yes | |
| POST | `/api/spaces/{spaceId}/members` | yes | `{ email, role }` |
| PUT | `/api/spaces/{spaceId}/members/{memberId}` | yes | `{ role }` |
| DELETE | `/api/spaces/{spaceId}/members/{memberId}` | yes | |

Roles are `owner`, `admin` and `member`.

## Clients

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/api/clients` | yes | `{ name, email, company, phone, notes }` |
| GET | `/api/clients` | yes | |
| GET | `/api/clients/{clientId}` | yes | |
| PUT | `/api/clients/{clientId}` | yes | |
| DELETE | `/api/clients/{clientId}` | yes | |

## SMTP

One configuration per user. The password is never returned.

| Method | Path | Auth | Notes |
|---|---|---|---|
| GET | `/api/smtp` | yes | `{ host, port, username, from_email, from_name, updated_at }` |
| PUT | `/api/smtp` | yes | `{ host, port, username, password, from_email, from_name }` |
| DELETE | `/api/smtp` | yes | |
| POST | `/api/smtp/test` | yes | `{ to }`, sends a test message |

## Webhooks

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/api/webhooks` | yes | `{ url, secret }` |
| GET | `/api/webhooks` | yes | Secrets are not returned |
| GET | `/api/webhooks/{id}` | yes | |
| PUT | `/api/webhooks/{id}` | yes | `{ url, secret, enabled }` |
| POST | `/api/webhooks/{id}/test` | yes | Sends a sample payload |
| DELETE | `/api/webhooks/{id}` | yes | |

### Outgoing payloads

Plume POSTs JSON with `User-Agent: Plume-Webhook/1.0` and
`x-plume-signature-256: sha256=<hex>`, the HMAC-SHA256 of the raw body keyed on the
webhook's secret. Verify it before trusting anything in the body.

```json
{
  "event_id": "…",
  "event_type": "signer.signed",
  "occurred_at": "2026-01-01T12:00:00Z",
  "owner": { "id": 1, "name": "…", "email": "…" },
  "document": { "id": 7, "name": "…", "status": "pending", "file_name": "…",
                "url": "…", "sequential": false,
                "created_at": "…", "updated_at": "…" },
  "signer": { "id": 3, "name": "…", "email": "…", "role": "signer",
              "status": "signed", "order_num": 0, "signing_url": "…",
              "signed_at": "…", "viewed_at": "…", "email_opened_at": "…" }
}
```

`signer` is absent on document-level events. Event types: `document.created`,
`document.sent`, `document.completed`, `document.declined`, `document.deleted`,
`signer.added`, `signer.email_opened`, `signer.viewed`, `signer.signed`, `signer.declined`,
`signer.reminded`.
