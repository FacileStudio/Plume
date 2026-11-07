# Changelog

## 0.10.1 - 2026-11-07

### Changed

- One user menu and one settings shell (`@facile/muse` v0.13.0 to v0.14.1). The sidebar card and the
  mobile avatar open the muse `UserMenu`: account, settings, language, theme, help and sign out.
  The cog link is gone. Settings use `SettingsLayout` in the canonical order: Profile, Preferences,
  Notifications, Integrations, API tokens, Advanced. Profile is read-only and links to Facile, so
  the name and email form and the in-settings sign-out are removed. The clear-local-data action
  sits in a `DangerZone`.
- The API tokens section is muse `TokensSection`. The API calls and token semantics are
  unchanged (a token is created from a label, bearer auth, no expiry, so the expiry choice in the
  drawer is not sent). The API reference rows stay in that section.
- `appearance` is now `preferences` and `api` is now `tokens`; the old addresses redirect.
  `/settings` shows the section list on a phone.
- A language picked in the menu is stored in the `fc_locale` cookie and wins over the Registre
  profile locale on the next visit.

## 0.10.0 - 2026-11-07

### Changed

- Libraries: porte v0.5.2 to v0.12.0, porte/directory v0.1.0 to v0.3.0, tronc v0.16.0 to v0.17.1,
  `@facile/muse` v0.11.0 to v0.13.0.
- Sign-in goes through Registre only. With an issuer configured porte mounts no password route, and
  `OIDC_ISSUER` is now required when `APP_ENV=production`. The login page is muse `AuthScreen`, and
  the landing page uses `LandingHeader`, `LandingFooter` and `SignInWithFacile`.
- `UserStore.UpsertFromOIDC` links an existing account only by a verified email and refuses when the
  account already holds another subject of the same issuer (porte v0.10.0 rules). It stores the
  `locale` claim.
- The UI is available in fr, en, es and de. The locale comes from the Registre profile, then the
  `fc_locale` cookie, then the browser. Dates and numbers use `Intl`. es and de need native review.
- tronc security headers replace the local middleware. The client's inline scripts are hashed into
  `script-src`, and `connect-src` allows Ardoise's plan catalog. The full CSP stays report-only.
- `APP_VERSION` build arg, when set, is the version `/version` reports; otherwise the commit.
- The mise Go version and CI use `go.mod`'s Go 1.26. CI now runs build, tests against Postgres, the
  message check, `svelte-check` and filet.

### Added

- GDPR deletion: `POST /api/registre/deletions` and a poll of Registre's deletion list erase the
  person's account, personal documents with their files, contacts, webhooks, SMTP settings and
  memberships. Documents the person created in a space others use stay and pass, with the space, to its
  longest-standing admin, else its longest-standing member, so they stay editable and keep
  archiving to Nuage. Idempotent.
- Migrations (idempotent, run at boot): `users.locale`, `porte_login_codes.provider` and `.subject`,
  tables `directory_deletion_state` and `porte_magic_links` (porte's erase touches it). `porte_identities` token columns are untouched, so the Nuage
  app-folder delegation keeps working.
- `locale` in `GET /api/auth/me`.

### Removed

- `POST /api/auth/register`, `POST /api/auth/login`, `PUT /api/auth/password`, the register and
  password UI, and `SSO_ONLY`. API tokens, signing links and webhooks are unchanged.

## 0.9.0 - 2026-11-07

### Changed

- Signed PDFs are archived in the document owner's private Plume folder in Nuage (app folders)
  instead of a Nuage space. Plume refreshes the owner's stored Registre token, exchanges it for a
  5-minute `nuage.write` token and calls `PUT /api/app/object`. An owner with no stored Registre
  refresh token is skipped and logged.
- `NUAGE_AUDIENCE` (Nuage's Registre client id) replaces `NUAGE_API_TOKEN`,
  `NUAGE_DEFAULT_SPACE_FACILE_ID` and `NUAGE_DEFAULT_SPACE_NAME`. `NUAGE_URL` now also needs
  `OIDC_ISSUER`.
- `nuage_space_facile_id` on signature requests is still accepted and echoed but selects nothing.

### Removed

- The find-or-create Nuage space provisioning.

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
