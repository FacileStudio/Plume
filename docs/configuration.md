# Plume — Configuration

Every environment variable the API actually reads, taken from `internal/env/env.go` and the
`tronc/env` and `tronc/spa` packages it builds on.

## Core

These come from `troncenv.LoadCore()`, shared by every Go app in the suite.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `DATABASE_URL` | yes | — | Postgres connection string. `LoadCore` returns an error without it and the process exits 1 |
| `PORT` | no | `8080` | HTTP listen port. Must parse as an integer and land in 1–65535, or startup fails |
| `LOG_LEVEL` | no | `info` | One of `debug`, `info`, `warn`, `error`. Anything else fails startup |
| `APP_ENV` | no | `development` | `development`, `staging` or `production`. Never gates security behavior |
| `CORS_ALLOWED_ORIGINS` | no | — | Comma-separated allowed origins. An unset list allows no cross-origin caller |
| `JOURNAL_URL` | no | — | Journal ingest URL. Log shipping only turns on when both this and the token are set |
| `JOURNAL_TOKEN` | no | — | Per-app Journal key |

`CORS_ALLOWED_ORIGINS` is the canonical name. `tronc/env` falls back, in order, to
`ALLOWED_ORIGINS`, `DOMAINS`, `DOMAIN`, `CORS_ORIGINS`, `TRUSTED_ORIGINS` and
`CLIENT_ORIGIN`, so an unmigrated deployment keeps working. In practice this means setting
`DOMAIN` alone also sets the CORS origin.

## Plume

| Variable | Required | Default | What it does |
|---|---|---|---|
| `DOMAIN` | no | `http://localhost:5173` | Public app URL. Used to build signing links in emails, webhook payload URLs, and the default OIDC success redirect |
| `UPLOAD_DIR` | no | `/data/uploads` | Root for uploaded PDFs, generated certificates, audit trails and avatars. `UPLOAD_DIR/avatars` is created at startup |
| `SSO_ONLY` | no | `false` | When true, `/api/auth/register` and `/api/auth/login` are never registered. Must parse as a Go boolean |
| `CLIENT_DIR` | no | `./client` | Directory holding the built SPA. Read by `tronc/spa`. The Dockerfile pins `/client` explicitly, because the distroless base carries its own working directory and a relative path would resolve elsewhere |

The default `DOMAIN` points at the Vite dev server, which is right for `go run .` next to
`bun run dev` and wrong for anything else. Compose sets it.

## OIDC

OIDC is off until `OIDC_ISSUER` is set. Once it is, three more variables become required and
`env.Load` fails without them, taking startup down with it.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `OIDC_ISSUER` | no | — | Discovery URL, e.g. `https://porte.facile.studio/application/o/plume/`. Setting it enables SSO |
| `OIDC_CLIENT_ID` | with issuer | — | Client ID from the provider |
| `OIDC_CLIENT_SECRET` | with issuer | — | Client secret |
| `OIDC_REDIRECT_URL` | with issuer | — | Must point at `<DOMAIN>/api/auth/oidc/callback` |
| `OIDC_SUCCESS_URL` | no | value of `DOMAIN` | Where the callback redirects the browser, with `?code=` appended |

Requested scopes are `openid`, `email`, `profile` and `offline_access`. The provider must
return an `email` claim and a truthy `email_verified`, or the callback rejects the login.

## Compose-only variables

`docker-compose.yml` reads these to configure the `plume-db` service. The API never sees
them.

| Variable | Default | What it does |
|---|---|---|
| `POSTGRES_USER` | `postgres` | Database superuser |
| `POSTGRES_PASSWORD` | `postgres` | Its password |
| `POSTGRES_DB` | `plume` | Database name |

Note that the compose file hardcodes `DATABASE_URL` to
`postgres://postgres:postgres@plume-db:5432/plume?sslmode=disable`, so changing the three
variables above without changing that string will break the connection.

## SMTP

SMTP is not configured through the environment. Each user stores their own host, port,
username, password, from-address and from-name in the `smtp_configs` table, through
`/api/smtp`. A user with no SMTP config cannot send signing invitations.

## The failure mode worth knowing

Every configuration error is fatal and silent-ish: `env.Load` returns an error, `run()` logs
`failed to load config` and returns, and `main` calls `os.Exit(1)`. A container that restarts
in a loop with one log line is almost always a missing `DATABASE_URL`, a malformed `PORT`, an
unrecognized `LOG_LEVEL`, or an `OIDC_ISSUER` set without its three companions.
