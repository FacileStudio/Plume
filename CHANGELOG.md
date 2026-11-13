# Changelog

## Unreleased

## 0.14.1 - 2026-11-13

### Documentation

- `docs/architecture.md` names the 23 production signer rows written before v0.13.4 whose `ip_address` holds a proxy address instead of the signer's, explains what that address means, and records why the rows and their PDFs stay unchanged.

## 0.14.0 - 2026-11-13

### Changed

- Database: the schema is ordered goose SQL under `apps/api/migrations/`, applied at boot by `tronc/migrate` behind a Postgres advisory lock. GORM `AutoMigrate` and the hand-written `CREATE TABLE IF NOT EXISTS` and `CREATE INDEX IF NOT EXISTS` statements are gone from the boot path. The GORM structs stay as the query layer: `TestGORMModelsMatchTheMigratedSchema` fails when a struct names a column the migrations do not create, or a type its field cannot hold. `scripts/check.sh` fails on an `.AutoMigrate(` call outside a test file.
- Database: `00001_baseline.sql` is production's schema as `pg_dump --schema-only` printed it on 2026-11-13, including porte's tables, `directory_deletion_state` and tronc/mcp's `mcp_actions`. On its first boot, a database built before this release (Plume's tables, no applied migration) is stamped at version 1 with tronc/migrate's `baseline` command, so the baseline is recorded without running. A database built from the migrations dumps identical to production. `apps/api/schema.sql` is the normalized dump of the whole set. `TestThePorteTablesMatchPortePG` and `TestTheMCPSchemaIsInTheMigrations` fail when a library bump changes those tables.
- Boot: the data steps that ran at every boot still run at every boot, after the migrations and in the same order (`schemas.Reconcile`): the space `facile_id` backfill, the avatar column backfill, the porte local identity re-key, the pre-porte session carry-over, the password and OIDC subject adoption, and the Registre picture backfill. The `DROP INDEX IF EXISTS idx_signers_token` before AutoMigrate, which then rebuilt the same non-unique index, is gone: the baseline holds the index. The document hash backfill is unchanged.
- build: `github.com/FacileStudio/tronc/migrate` v0.1.1 with `pressly/goose/v3` v3.27.3; goose raises `segmentio/asm` from v1.1.3 to v1.2.1.

### Added

- `docker run <image> migrate status` (also `up`, `version`, `baseline <n>`) answers without a shell.
- Database: `00002_column_defaults.sql` gives every `created_at`, `updated_at` and `joined_at` a `now()` default, and gives the users columns porte/kit does not write the values the kit's `OnCreate` writes (`''` and `0001-01-01 00:00:00+00`). A row inserted outside GORM no longer gets NULL there. It has a Down.

### Rollback

- Rolling back to v0.13.x takes the 00002 defaults off again, because AutoMigrate removes every default its struct tags do not declare, while `goose_db_version` still records version 2. The rehearsal (old, new, old, new binary on a copy of production's schema) showed exactly that. Before deploying this release or a later one after such a rollback, run `DELETE FROM goose_db_version WHERE version_id = 2`. Accounts created in between are still filled, because `OnCreate` writes the same values.

## 0.13.4 - 2026-11-12

### Fixed

- The audit trail records the signer's real IP. The sign and decline routes read `X-Real-IP` by hand, which behind Cloudflare and Traefik holds the Cloudflare edge, and from a direct caller holds whatever it sent. They now take the host of `RemoteAddr` as tronc's real-IP middleware resolved it (`TRUSTED_PROXIES=private,cloudflare` in production), so a spoofed header from an untrusted peer is ignored and the port is no longer stored.

## 0.13.3 - 2026-11-12

### Fixed

- An account re-created for a person Registre already erased (a bearer token minted before the erasure, an MCP first call, a late snapshot) is erased again on the next deletion push or poll, with no restart. Sign-in, MCP provisioning and the profile sync refuse a subject Plume erased, and the Registre snapshot sync no longer provisions one. porte v0.19.0 (`Deletions.Holds`, `oidc.Deps.Erased`), porte/directory v0.5.0.

## 0.13.2 - 2026-11-11

### Fixed

- The favicon, touch and manifest icons, link preview image and web manifest come from the muse v0.19.0 brand generator, and the head carries the suite's canonical title, description, theme colours and Open Graph and Twitter tags. The landing page title and description match the suite catalog.

## 0.13.1 - 2026-11-11

### Security

- Client dependencies past their high advisories: `pdfjs-dist` 6.5.438, `@sveltejs/kit` 2.70.4 and `vite` 7.3.8, with overrides pinning `devalue` 5.9.4, `nanoid` 3.3.20, `postcss` 8.5.32 and `source-map-js` 1.2.2.

### Fixed

- The signing page opens the PDF with `getDocument({ url })`. pdf.js 6 refuses the bare URL string the page passed before.

## 0.13.0 - 2026-11-11

### Fixed

- `DELETE /signers/{id}` deletes the signer's fields in the same transaction. Before, the fields stayed on the draft, pointing at a signer that no longer existed.

### Added

- MCP tool `plume_remove_signer`: takes a person off a draft the user owns, with their fields, through the same service call as `DELETE /signers/{id}`. A normal write, not confirmed; a document already sent keeps its signers.

## 0.12.0 - 2026-11-11

### Added

- **Remote MCP server at `/mcp`**, on tronc/mcp v0.22.1 and porte/mcpauth v0.2.1, behind Registre access tokens bound to `MCP_RESOURCE` (scopes `plume.read` and `plume.write`). Any MCP client (Claude, ChatGPT, Cursor) connects after a Registre sign-in, free on every plan; the first call creates the Plume account as a first web sign-in would. Ten tools: `plume_list_documents`, `plume_get_document`, `plume_list_clients`, `plume_create_document` (a base64 PDF of up to 3 MB as a draft with its signers), `plume_update_document`, `plume_add_signer`, `plume_save_client`, and three the person always confirms: `plume_send_document`, `plume_remind_signer` and `plume_delete_document`. A client that can ask confirms in its own UI (MRTR); any other gets a link to the new `/pending` page. Every tool calls the service behind the matching REST route, so spaces and ownership hold as in the web app, and signing links never reach the model.
- `mcp_actions` table (tronc/mcp's additive DDL) for the actions waiting on `/pending`; erasing a person deletes theirs.

### Changed

- tronc v0.21.0 to v0.22.1, `tronc/testdb` v0.1.2, porte v0.18.2 to v0.18.4, `modelcontextprotocol/go-sdk` v1.8.0.

## 0.11.0 - 2026-11-10

### Changed

- Licensed under the GNU Affero General Public License v3.0 only (`AGPL-3.0-only`). See `LICENSE`. Releases up to 0.10.5 carried no licence.
- **Erasure keeps the legal records.** When Registre erases a person, every document they sent for signature or signed stays byte-identical, with its signers, fields, files and audit trail. Before, their personal documents were deleted and their space documents passed to another owner. The account row, sessions, API tokens, SMTP settings, webhooks, unsent drafts and contacts no sent document names are deleted. Completed and declined documents get their audit trail written to disk first, so it keeps the owner's name and email. Field by field: `docs/erasure.md`.
- **The verify routes answer `rate_limited`** instead of `resource_exhausted` on 429, with the real wait in `Retry-After` instead of a fixed 60. The limit is tronc's `middleware.Ratelimit`: a burst of 10, then one every 3 seconds, so at most 30 in any minute (the old bucket could admit 40), keyed on the visitor address, with a memory bound.
- **Sign-in goes through `porte/kit`.** The stored email changes only when Registre marks the new one verified (suite rule). New accounts get the same column values GORM wrote.
- The browser uses porte's session cookie only. The `plume.token` and `token` keys older builds kept in `localStorage` are deleted on load. API tokens for Ardoise are unchanged.
- tronc v0.17.1 to v0.21.0, porte v0.13.0 to v0.18.2, `@facile/muse` v0.16.1 to v0.18.0, `@facile/client` v0.1.1, Go 1.26.11, `golang.org/x/crypto` v0.58.0, `golang.org/x/image` v0.48.0, `golang.org/x/text` v0.44.0.
- CI: filet v0.19.1 with `preset: facile`, douane v0.7.1 (`-fail critical -scope prod`), and `filet gate .` before push through `lefthook.yml`.

### Added

- The user menu shows pending Registre space invitations (`GET /api/me/invitations`).

## 0.10.5 - 2026-11-07

### Changed

- `@facile/muse` v0.14.4 to v0.16.1. Secondary buttons use the `secondary` variant in place of the deprecated `outline`, and the remove buttons of the signer and field lists use `ghost-danger`. The icon generator accepts Iconify aliases, which the new muse icons need.
- `GET /version` comes from the version in `apps/client/package.json`, which the release commit bumps (still `v` plus the number). The `APP_VERSION` build arg and the Dokploy env no longer take part, so a stale env value cannot make the server report an old release. The image build no longer needs `.git`.

## 0.10.4 - 2026-11-07

### Changed

- `@facile/muse` v0.14.2 to v0.14.4: the language chosen in the user menu re-renders the whole UI without a reload, and a single 100% slice in the dashboard donut is drawn as a clean ring.
- The user menu's Subscription row drops its empty `plan` placeholder and shows only its label.

## 0.10.3 - 2026-11-07

### Security

- Registre tokens are sealed at rest. porte v0.12.0 to v0.13.0: `porte_identities.access_token` and
  `refresh_token` (the refresh token Plume exchanges for a Nuage write token) are encrypted with
  AES-256-GCM under `PORTE_TOKEN_KEY`. Existing rows are sealed once, in the background, after boot,
  and the count is logged. An invalid `PORTE_TOKEN_KEY` stops the boot. With the key unset nothing changes.
- Keep `PORTE_TOKEN_KEY` once set. Losing it makes the stored tokens unreadable: people sign in again and
  nothing else is lost. Do not roll back to a build older than porte v0.13.0 after the first reseal.

## 0.10.2 - 2026-11-07

### Changed

- `@facile/muse` v0.14.1 to v0.14.2: the user menu clamps itself inside the viewport at the right edge of a header.
- The user menu has a Subscription row that opens `https://sso.facile.studio/spaces`, where plans are managed.
- API tokens: the create drawer has no expiry field and the list has no expiry column, because Plume tokens never expire.

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
