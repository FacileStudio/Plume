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
mise run hooks         # git config core.hooksPath .githooks
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

Migrations run automatically on every start through `schemas.Migrate`, which is GORM
`AutoMigrate` plus two hand-written index statements. There is no migration file to write
and no down migration — adding a column means adding a struct field.

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

`.githooks/pre-push` runs `scripts/check.sh --go-only`, not the full gate. Plume's client
carries pre-existing `svelte-check` errors, so gating on the full check would block every
push until those are cleared. Run `sh scripts/check.sh` yourself for the whole picture. When
the client reaches zero errors, drop the `--go-only`.

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
  through `dynamicCompileOptions` in `svelte.config.js`. UI primitives under
  `src/lib/components/ui/` are shadcn-svelte managed; do not hand-edit them.
