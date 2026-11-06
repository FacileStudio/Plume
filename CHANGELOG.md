# Changelog

## 0.8.0 - 2026-11-06

### Added

- The landing page shows Facile's plans under "Pricing", from Ardoise's public catalog. The
  section is absent while the catalog is empty or unreachable.

### Changed

- `@facile/muse` v0.6.6 to v0.11.0.

## 0.7.1 - 2026-11-03

### Changed

- Spaces use the share-circle icon, the one every Facile app now shows for spaces.

## 0.7.0 - 2026-11-03

### Added

- Shared spaces come from Registre. Plume hands Registre its existing spaces once, merged by
  name, then keeps every space you are in, with its members and roles, equal to Registre's
  within seconds of a change and at least every five minutes. Someone added in Registre is
  provisioned before their first sign-in.
- `POST /api/registre/sync` is the nudge Registre posts after a change. It needs no
  credentials and only schedules an authenticated pull.

### Changed

- With SSO configured, creating, renaming, deleting a space, adding, removing or re-ranking a
  member and leaving are done in Registre. Plume shows a "Manage in Registre" link and the API
  answers 409 for a linked space. Member lists stay, read-only. Local password accounts in a
  space are never removed by a sync. Without SSO nothing changes.

## 0.6.0 - 2026-11-03

### Added

- Space members and teammate signers show their suite profile picture, the one set in Registre.
  A change there shows in Plume on the next page load. External signers keep their initials.
- Boot fills the picture URL for SSO users who have not signed in since, so nobody has to sign
  in again to get theirs.

### Changed

- `@facile/muse` moves to v0.6.5. A missing picture falls back to the initial instead of a
  broken image.

### Security

- `/api/files/*` no longer exists. It served the whole upload volume with directory listings,
  so every stored PDF, certificate and audit trail was readable without a session. The legacy
  uploaded-avatar fallback it existed for is gone; Plume had no upload form left.

## 0.5.0 - 2026-10-31

### Added

- `POST /api/documents/signature-requests` takes an optional `client` (`name`, `email`,
  `company`). The document is filed under the caller's client with that email, which is
  created when missing. Clients sent by Ardoise no longer leave Plume documents unfiled.

## 0.4.0 - 2026-10-31

### Added

- Settings > API generates, lists and revokes named API tokens. A token is shown once, acts as
  the account, has no expiry and survives a sign-out.
- `GET /api/auth/api-tokens` lists the caller's tokens (label, creation, last use; never the
  value) and `DELETE /api/auth/api-token/{id}` revokes one. A login session answers 404 there.

### Fixed

- The API page claimed Plume issues no API keys and told people to script against a dedicated
  account. It now matches `POST /api/auth/api-token`.

## 0.3.0 - 2026-10-30

### Added

- `GET /api/version` (and `/version`) reports the commit the image was built from, or `dev`.
  The Dockerfile reads it from `.git`, so `.dockerignore` now keeps `.git`.
- Link with Ardoise (see `docs/ardoise-link.md`): `POST /api/documents/signature-requests`
  creates a document from a PDF with its signers and fields and sends it in one call;
  `POST /api/auth/api-token` mints a long-lived bearer token for the calling app;
  `GET /api/documents/lookup?external_ref=` reads a document's status by the caller's own
  reference.
- Documents take three optional fields: `external_ref`, `nuage_space_facile_id` and
  `nuage_linked_to`, plus a `nuage_published_at` set by the archive. The migration is additive;
  a partial unique index on `(owner_id, external_ref)` makes a retried request return the
  existing document.
- The Nuage archive honours `nuage_space_facile_id` and `nuage_linked_to` when set.
- Webhook payloads carry `external_ref` and a `nuage` object, and a new `document.archived`
  event follows a successful Nuage archive.

### Fixed

- `docker-compose.yml` and `compose.yml` now forward `NUAGE_URL`, `NUAGE_API_TOKEN`,
  `NUAGE_DEFAULT_SPACE_FACILE_ID` and `NUAGE_DEFAULT_SPACE_NAME` (empty by default). Until
  now the container never received them, so the v0.2.0 archive could not run in production.

## 0.2.0 - 2026-10-30

### Added

- Completed documents can be archived in Nuage. When the last signer finishes, the stamped PDF
  is published in the background to the Nuage space that shares the Plume space's `facile_id`
  (Nuage creates it on first use), or to a default space for documents with no space. It is
  off unless `NUAGE_URL` and `NUAGE_API_TOKEN` are set, retries up to five times with backoff,
  runs at most four jobs at once, and never blocks or fails signing. The publish carries
  `plume:document:<id>`, so a retry never stores a second copy.
- Spaces have a `facile_id`, the suite-wide key. Existing spaces get one at migration, and a
  unique index enforces it.

## 0.1.0 — 2026-09-20

### Changed

- filet.yml harmonised to the strict suite standard, and the API split to meet
  it: per-route registrars, per-route documentation tables, the documents
  upload machinery in `upload.go`, the spaces member routes in `members.go`,
  and the pdfutil field painters in `stamp.go`
- dev compose: new `docker-compose.dev.yml` overlay publishes the database
  port on the loopback interface only

### Fixed

- pdfutil: restored the `FlattenFields` doc comment mangled by a merge, and
  the overlay temp file is now removed on every error path
- urlsafe: the SSRF guard no longer re-parses its CIDR table per probe; the
  table is covered by a regression test
- auth: the documentation route table appends in the non-nested form
