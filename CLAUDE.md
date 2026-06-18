# Plume

Self-hosted document signing platform. DocuSeal alternative, single-tenant.

## Tech Stack

- **API**: Go 1.24, Chi router, GORM + PostgreSQL 16, fpdf for PDF generation
- **Client**: SvelteKit (Svelte 5 runes), TypeScript, Tailwind v4, shadcn-svelte (nova style), pdfjs-dist
- **Auth**: Session tokens (Bearer), OIDC/SSO optional (Authentik-compatible)
- **Runtime**: Bun (client), distroless container (API)
- **Deploy**: Docker Compose (3 services: `plume-db`, `plume-api`, `plume-client`)

## Project Structure

```
apps/
  api/                  Go API server (port 4000)
    main.go             Entrypoint, router setup, graceful shutdown
    internal/           Shared infra (middleware, env, database, httpjson, pdfutil, hashing)
    modules/            Domain modules (auth, documents, signers, fields, signing, smtp, webhooks, verify, reminders)
    schemas/            GORM models and migrations
    Dockerfile          Multi-stage: golang:1.24-alpine -> distroless
  client/               SvelteKit frontend (port 3000)
    src/lib/backend.ts  API client with all typed endpoints
    src/lib/components/ Field editor, PDF viewer, signature pad, shadcn-svelte UI
    src/routes/         Pages: (app)/{dashboard,documents,profile,settings}, login, share/[token], verify
    hooks.server.ts     Reverse proxy: /api/* -> plume-api:4000
    Dockerfile          Multi-stage: oven/bun
docker-compose.yml      Full stack orchestration
.env.example            All env vars documented (all optional, defaults work)
```

## Key Commands

### API (`apps/api/`)

```bash
go run .                        # Run API server locally
go build -o bin/api .           # Build binary
go vet ./...                    # Lint
go test ./...                   # Run tests
```

### Client (`apps/client/`)

```bash
bun install                     # Install dependencies
bun run dev                     # Dev server (Vite, port 5173)
bun run build                   # Production build
bun run preview                 # Preview production build
bun run check                   # Svelte type checking
```

### Docker (project root)

```bash
docker compose up --build       # Full stack (db + api + client)
docker compose down             # Stop all services
```

## Environment Variables

All optional with sane defaults. See `.env.example` and `apps/api/.env.example`.

| Variable | Default | Notes |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@db:5432/plume?sslmode=disable` | Postgres connection string |
| `PORT` | `4000` | API listen port |
| `DOMAIN` | `http://localhost:5173` (dev) / `http://localhost:3000` (compose) | Used for CORS and email links |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `UPLOAD_DIR` | `/data/uploads` | PDF storage directory |
| `API_URL` | `http://localhost:4000` | Client-side: where the SvelteKit proxy sends requests |
| `OIDC_*` | unset | OIDC/SSO config (all four required if `OIDC_ISSUER` is set) |
| `SSO_ONLY` | `false` | Disable local email/password auth |

## Architecture Notes

- The SvelteKit client proxies all `/api/*` requests to the Go API via `hooks.server.ts`. The client never talks to the API directly from the browser.
- Auth is Bearer token in localStorage, passed through the proxy.
- Document workflow: `draft` -> `pending` -> `completed` / `declined`.
- Modules are self-contained: each has routes, service, and handlers. Subrouters are composed in `main.go` via `RegisterRoutes` + `*Routes` helper functions.
- Migrations run automatically on startup via `schemas.Migrate(db)` (GORM AutoMigrate).
- PDF uploads stored on filesystem (`UPLOAD_DIR`), persisted via named Docker volume `plume_uploads`.
- Webhook dispatch uses HMAC-SHA256 signing.
- Reminder ticker runs in a background goroutine with configurable per-user interval.
- The client uses shadcn-svelte (nova style, neutral base color) with Lucide and Iconify icons.

## Conventions

- Go code follows standard `internal/` + `modules/` layout. No framework beyond Chi + GORM.
- Svelte 5 runes API (`$state`, `$props`, `$derived`, `$effect`) enforced via `dynamicCompileOptions` in svelte.config.js.
- UI components live in `src/lib/components/ui/` (shadcn-svelte managed).
- API responses use `httpjson.WriteJSON` for consistent JSON output.
- No test framework is set up on the client side yet.
