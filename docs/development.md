# Plume — Development

Local setup, the two processes you run, the tests, and the quality gate that guards pushes.

## Prerequisites

| Tool | Version | Why |
|---|---|---|
| Go | 1.25 | `apps/api/go.mod` declares `go 1.25.0`; `mise.toml` pins the toolchain |
| Bun | 1.x | Client package manager and build runner |
| PostgreSQL | 16 | What the Compose stack and production run |
| mise | any | Task runner for `install`, `check`, `format`, `hooks` |
| Docker | any | Only if you want the Compose stack instead of a local Postgres |

## Setup

```sh
mise run install       # bun install --frozen-lockfile in apps/client
mise run hooks         # lefthook install --reset-hooks-path
```

The second one is not optional in spirit — the pre-push hook is the only thing standing
between a broken build and `main`.

Bring up a database. The quickest path is the Compose Postgres alone:

```sh
docker compose up -d plume-db
```

## Running

Two processes, two terminals.

```sh
cd apps/api
DATABASE_URL=postgres://postgres:postgres@localhost:5432/plume?sslmode=disable go run .
```

```sh
cd apps/client
bun run dev            # Vite on http://localhost:5173
```

The API listens on `:8080` unless `PORT` says otherwise, and serves no SPA in this mode —
`tronc/spa` only mounts the catch-all when `CLIENT_DIR` actually contains a build. The Vite
dev server talks to the API cross-origin, which is why the default `DOMAIN` is
`http://localhost:5173`: it doubles as the CORS origin.

Migrations run automatically at startup (`schemas.MigrateWithIssuer`). To change the schema,
add the next numbered file under `apps/api/migrations/` with both `-- +goose Up` and
`-- +goose Down`, update the struct in `schemas/` by hand, and regenerate `apps/api/schema.sql`
from a database the migrations built:

```sh
PGTZ=UTC pg_dump --schema-only --no-owner --no-privileges --no-comments --no-tablespaces --schema=public \
  --exclude-table=goose_db_version --exclude-table=goose_db_version_id_seq "$DATABASE_URL" \
  | grep -vE '^(SET |SELECT pg_catalog\.set_config|\\restrict|\\unrestrict|--|CREATE SCHEMA)' | cat -s
```

Never edit a migration that has shipped: the ledger records it as applied and the change never
runs. The test schemas persist between runs for the same reason, so after editing an unshipped
file, drop the `plume_*` schemas in the test database. `go run . migrate status` lists what is
applied. `scripts/check.sh` fails on any `.AutoMigrate(` outside a test file.

## Tests

```sh
cd apps/api && go test ./...
```

Five packages carry tests: `internal/documentation` (OpenAPI assembly),
`modules/reminders`, `modules/signers`, `modules/signing` (timestamp formatting) and
`modules/verify`. They are plain `go test`, no fixtures and no database. The client has no
test framework.

## The quality gate

`scripts/check.sh` is the gate. It depends on nothing but `go` and, for the client half,
`bun`. It is deliberately a shell script rather than a mise task body: `mise run` resolves
every tool in the merged config before running anything, so one broken tool in your global
config would otherwise take the gate down with it.

```sh
sh scripts/check.sh              # gofmt -l, go vet, go test, then the client type-check
sh scripts/check.sh --go-only    # Go half only
sh scripts/check.sh --format     # rewrite Go sources in place
```

Equivalent mise tasks: `mise run check`, `mise run check-go`, `mise run format`.

The script resolves `go` and `gofmt` from `GOROOT` when it is set. mise exports `GOROOT` for
the pinned version but leaves an unrelated `go` earlier on `PATH` — Homebrew's, typically —
and mixing the two produces `compile: version "X" does not match go tool version "Y"`.

If `bun` is missing, the client half is skipped with a warning rather than failing.

## The pre-push hook

`lefthook.yml` runs `filet gate .` before every push: gofmt, `filet check` (`preset: facile`
in `filet.yml`), `go vet`, staticcheck, svelte-check, the Go and client tests, and
`douane scan -fail critical -scope prod`. The Go tests need `TEST_DATABASE_URL`. A finding
that no released version fixes yet gets a dated permit in `douane.yml`, never a pin to an
unreleased commit. `mise run hooks` installs the hook; the shared commit-msg rules come from
`FacileStudio/hooks`.

Bypass once with `git push --no-verify`.

## Conventions

- Go layout is `internal/` for shared infrastructure and `modules/` for domain code. Each
  module owns `router.go`, `service.go` and `types.go`; nothing beyond Chi and GORM.
- New routes are registered in the module's `RegisterRoutes` and composed in `main.go`.
  Routes that belong under another module's path go in a `*Routes` helper passed as a
  nested function.
- Handlers return through `httpjson.WriteJSON` and `httpjson.WriteError` from tronc, so
  error shapes stay uniform.
- The client is Svelte 5 runes only — `$state`, `$props`, `$derived`, `$effect` — enforced
  through `dynamicCompileOptions` in `svelte.config.js`.
- UI comes from `@facile/muse`, pinned to a tag. Compose its components rather than writing
  markup, and read `CHARTE.md` plus `demo/src/pages/` in the muse checkout before building a
  page — the demo is the visual spec, the charte is the prose. The two local components that
  remain (`field-editor`, `signature-pad`) exist because muse has no PDF field placer and no
  signature canvas.
- After adding an icon anywhere, or after bumping `@facile/muse`, run `bun run icons` in
  `apps/client`. It rescans both trees, refetches the collections and fails loudly on a glyph
  that does not exist — which is the point: an unbundled icon silently restores a runtime
  call to `api.iconify.design` from the user's browser.
- `vite dev` will not start without `optimizeDeps.exclude: ['@facile/muse']` in
  `vite.config.ts`. muse ships uncompiled `.svelte.ts` rune modules, and the dev-only
  dependency optimizer hands them to esbuild with no TypeScript transform. `vite build` never
  runs the optimizer, so the gate stays green while dev is completely broken.
