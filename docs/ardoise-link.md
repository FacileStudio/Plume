# Plume - Ardoise link contract

How another app (Ardoise) asks Plume to get a PDF signed, learns the result, and finds the
signed copy in Nuage. Every new field is optional. Existing callers see no change.

Base URL: the Plume host. All routes below live under `/api`.

## 1. Authentication

Plume authenticates with a bearer token. Mint a long-lived one once, as the Plume user that
will own the documents:

```
POST /api/auth/api-token
Authorization: Bearer <session token from /api/auth/login>
{"label": "ardoise"}

201 {"id": 12, "label": "ardoise", "token": "<shown once>"}
```

The token has no expiry. It acts as that user: it sees that user's documents and spaces, and
nothing else. Store it like a password. Revoke it by ending that session in Plume.

Every call below sends `Authorization: Bearer <token>`.

## 2. Create a document for signature

```
POST /api/documents/signature-requests
Content-Type: multipart/form-data
```

Two parts:

- `file`: the PDF (must start with `%PDF-`, 50 MB max).
- `request`: a JSON object, as text.

```json
{
  "name": "Contract 42 - Acme",
  "space_id": 3,
  "sequential": false,
  "send": true,
  "external_ref": "ardoise:contract:42",
  "nuage_space_facile_id": "fac_0123456789abcdef01234567",
  "nuage_linked_to": "ardoise:contract:42:signed",
  "signers": [
    {"name": "Ada Martin", "email": "ada@acme.example", "role": "signer", "order": 1},
    {"name": "Bob Lee", "email": "bob@acme.example", "role": "viewer"}
  ],
  "fields": [
    {"signer": 0, "field_type": "signature", "page": 2, "x": 8, "y": 70,
     "width": 30, "height": 8, "required": true, "label": "Signature"}
  ]
}
```

| Field | Required | Notes |
|---|---|---|
| `name` | yes | Document title |
| `signers` | yes | 1 to 20 entries. `name` and a valid `email` are required. `role` is `signer` (default), `approver` or `viewer`. `order` matters only when `sequential` is true |
| `space_id` | no | Plume space id. The caller must be a member |
| `sequential` | no | Default false. Signers sign in `order` |
| `send` | no | Default true: invitations go out in this call. False leaves a draft |
| `fields` | no | `signer` is the index of the signer in `signers`. `field_type` is `signature`, `text`, `date` or `checkbox`. `page` starts at 1. `x`, `y`, `width`, `height` are percentages of the page (0 to 100). When `fields` is empty, every `signer` and `approver` gets one required signature box on the last page |
| `external_ref` | no | Your reference. At most 200 characters, printable, no leading or trailing space. Unique per Plume user |
| `nuage_space_facile_id` | no | `fac_` plus 24 lowercase hex. Nuage space for the signed PDF |
| `nuage_linked_to` | no | Same validation as `external_ref`. Idempotency key for the Nuage upload |

Response, `201` when created:

```json
{
  "document": {
    "id": 57, "name": "Contract 42 - Acme", "status": "pending",
    "file_name": "contract.pdf", "owner_id": 1, "sequential": false,
    "original_hash": "...", "external_ref": "ardoise:contract:42",
    "nuage_space_facile_id": "fac_0123456789abcdef01234567",
    "nuage_linked_to": "ardoise:contract:42:signed",
    "created_at": "...", "updated_at": "..."
  },
  "signers": [
    {"id": 101, "name": "Ada Martin", "email": "ada@acme.example", "role": "signer",
     "status": "pending", "order_num": 1,
     "signing_url": "https://plume.example/share/<token>"}
  ]
}
```

Retries are safe. If `external_ref` already exists for this user, Plume creates nothing and
answers `200` with the existing document, in the same shape. The body of the retry is ignored.
Without `external_ref` every call creates a new document.

Errors use the usual shape; `400 invalid_argument` for any bad field, `403` for a space the
caller is not in. A failure after the document was stored removes it again, so a retry starts
clean.

`document.status` is `draft`, `pending`, `completed` or `declined`.

## 3. Read the status by external_ref

```
GET /api/documents/lookup?external_ref=ardoise:contract:42
```

Same response shape as above, always `200` or `404 not_found`. Scoped to the token's user.
`document.nuage_published_at` is set once the signed PDF is stored in Nuage. Signers carry
`status` (`pending`, `signed`, `declined`) and `signed_at`. Use it to poll when a webhook was
missed.

## 4. Completion webhook

Register once, with the same token:

```
POST /api/webhooks
{"url": "https://ardoise.example/hooks/plume", "secret": "<random, at least 32 bytes>"}
```

Always send your own `secret`. Plume signs with whatever is stored; an empty secret signs
with an empty key. Every event of every document owned by the token's user goes to every
enabled webhook, so filter on `event_type`.

Delivery: `POST` of a JSON body, `Content-Type: application/json`, `User-Agent: Plume-Webhook/1.0`.
Answer with any `2xx`. Plume tries 3 times (waits of 1 s, then 5 s, then 20 s) and then gives
up, which is why the status endpoint exists. Redirects are not followed.

### Signature

Header: `x-plume-signature-256: sha256=<hex>`

The value is the lowercase hex HMAC-SHA256 of the raw request body, keyed with the webhook
secret. Verify before parsing, compare in constant time:

```go
mac := hmac.New(sha256.New, []byte(secret))
mac.Write(rawBody)
want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
ok := hmac.Equal([]byte(want), []byte(r.Header.Get("x-plume-signature-256")))
```

There is no timestamp header. Deduplicate on `event_id` and treat the handler as idempotent.

### Events you need

`document.completed`: the last signer signed. Mark your record as signed on this one.

```json
{
  "event_id": "evt_9f2c...",
  "event_type": "document.completed",
  "occurred_at": "2026-10-30T10:00:00Z",
  "owner": {"id": 1, "name": "...", "email": "..."},
  "document": {
    "id": 57,
    "name": "Contract 42 - Acme",
    "status": "completed",
    "file_name": "contract.pdf",
    "url": "https://plume.example/documents/57",
    "sequential": false,
    "created_at": "...", "updated_at": "...",
    "external_ref": "ardoise:contract:42",
    "nuage": {
      "space_facile_id": "fac_0123456789abcdef01234567",
      "linked_to": "ardoise:contract:42:signed",
      "published": false
    }
  }
}
```

`document.archived`: the signed PDF is now in Nuage. Sent after `document.completed`, once the
upload succeeded (Nuage uploads retry up to 5 times, so this can lag or never arrive if Nuage
stays down; poll `nuage_published_at` then).

```json
{
  "event_type": "document.archived",
  "document": {
    "id": 57, "status": "completed", "external_ref": "ardoise:contract:42",
    "nuage": {
      "space_facile_id": "fac_0123456789abcdef01234567",
      "linked_to": "ardoise:contract:42:signed",
      "published": true
    }
  }
}
```

(Other document fields as above.) Rules for the `nuage` object:

- On `document.completed` it appears only when you set `nuage_space_facile_id` or
  `nuage_linked_to`, and echoes them with `published: false`.
- On `document.archived` it always appears and names the target actually used, with
  `published: true`. When you set nothing, that is the Plume space's `facile_id` (or the
  default space) and `plume:document:<id>`.
- `external_ref` is absent when you did not set one.

Other events (`document.created`, `document.sent`, `document.declined`, `signer.*`) also carry
`external_ref`. `document.declined` means the document will not be signed.

## 5. Where the signed PDF goes in Nuage

When the last signer completes, and only if Plume has `NUAGE_URL` and `NUAGE_API_TOKEN`:

| You set | Nuage space | Idempotency key (`linked_to`) |
|---|---|---|
| nothing | Plume space's `facile_id`, else `NUAGE_DEFAULT_SPACE_FACILE_ID`, else nothing is published | `plume:document:<id>` |
| `nuage_space_facile_id` | that space (Nuage creates it on first use) | `plume:document:<id>` |
| `nuage_linked_to` | as "nothing" | that key |
| both | that space | that key |

The two fields work independently. Re-publishing with the same key never stores a second copy.
The Nuage token Plume uses must be allowed to provision spaces (`provision: true`).

## 6. Summary of routes

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/auth/api-token` | Mint a long-lived bearer token |
| POST | `/api/documents/signature-requests` | Create, add signers and fields, send |
| GET | `/api/documents/lookup?external_ref=` | Status by your reference |
| POST | `/api/webhooks` | Register the callback URL and secret |
| GET | `/api/version` | Which build is running: `{"version":"<commit>"}` or `dev` |
