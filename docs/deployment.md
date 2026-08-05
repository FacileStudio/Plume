# Plume — Deployment

How the image is built, what Compose brings up, how Traefik routes to it, and what to watch
during an upgrade.

## The image

`Dockerfile` is a four-stage build producing one distroless image.

1. `oven/bun:1` installs the client dependencies from the lockfile and runs `bun run build`.
2. `golang:1.25-alpine` downloads modules, then builds a static binary with
   `CGO_ENABLED=0`, `-trimpath` and `-ldflags="-s -w"`.
3. A throwaway stage creates `/data/uploads` so the final image has the directory even
   before a volume mounts over it.
4. `gcr.io/distroless/static-debian12` receives the binary at `/api`, the built SPA at
   `/client`, and `ENV CLIENT_DIR=/client`.

That last line is load-bearing. The distroless base can carry its own working directory, so
a relative `./client` would resolve somewhere else entirely and the SPA would silently not be
served — the API would answer `/api/*` and 404 everything a human would type into a browser.

The image exposes `4000` and declares `/data` as a volume.

## Compose topology

Two services, matching the suite's one-container/one-router/one-hostname rule.

```
dokploy-network ──▶ traefik ──▶ plume-api:4000 ──▶ plume-db:5432
                                     │
                       uploads_data (/data/uploads)   db_data (/var/lib/postgresql/data)
```

| Service | Image | Notes |
|---|---|---|
| `plume-db` | `postgres:16-alpine` | `pg_isready` healthcheck every 5s, 10 retries |
| `plume-api` | built from `Dockerfile` | `expose: 4000`, no published port; depends on the database being healthy |

Named volumes are `plume_db_data` and `plume_uploads`. **Losing `plume_uploads` loses every
uploaded PDF** — the database only holds paths and hashes.

The default network pins the subnet `172.16.0.0/24`. Docker's default address pools
(`172.17`–`172.31/16` and `192.168.0.0/16` in `/20` blocks) are saturated on la ruche, so
letting Docker choose fails to create the project network at all.

## Traefik

Labels on `plume-api` declare two routers and one service, all on hostname
`plume.facile.studio`:

| Router | Entrypoint | Behavior |
|---|---|---|
| `plume-web` | `web` | Redirects to HTTPS through the `redirect-to-https@file` middleware |
| `plume-secure` | `websecure` | TLS with the `letsencrypt` cert resolver |

Both point at `plume-svc`, load-balancing to port `4000`. `traefik.docker.network` is
`dokploy-network`, which is declared external — Dokploy owns it.

One hostname serves everything. There is no separate API subdomain: `/api/*` and the SPA
share the origin, which is why the browser never has to make a cross-origin request in
production.

## Healthchecks

The container healthcheck is `/api healthcheck`, the binary invoking itself — `tronc`'s
`healthcheck.Handle` intercepts `os.Args` before anything else in `main`. That works in a
distroless image, which has no shell and no `curl`.

Over HTTP, `tronc/health` mounts four paths: `/health` and `/ready`, plus `/api/health` and
`/api/ready`. `/ready` pings the database; `/health` does not.

A green `/api/health` says the process is up. It says nothing about the SPA being served —
if `CLIENT_DIR` is wrong, health stays green while every page 404s. Check a real page, not
just the endpoint.

## Deploying to la ruche

Deployment is managed through Dokploy at `gare.facile.studio`, which builds from the repo
and runs the Compose file. Prefer the `dokploy` CLI over SSH and raw `docker`.

Environment comes from Dokploy's environment editor, not from a `.env` file in the repo.
`.env.example` is the template of what to paste there. At minimum, production needs `DOMAIN`
set to `https://plume.facile.studio`; without it, signing emails link to
`http://localhost:5173` and every invitation is dead on arrival.

## Migrations

There is no migration step to run. `schemas.Migrate` executes GORM `AutoMigrate` on every
boot, which means:

- Deploys are additive-safe. Adding a struct field adds a column.
- Removing or renaming a field does **not** drop or rename the column. The old column stays
  behind, unread.
- Two instances booting simultaneously both run `AutoMigrate`. Deploy one at a time.

A background goroutine also backfills `original_hash` for documents stored before hashing
existed. It logs `hash backfill complete` with a count, or warns and gives up. It reads every
unhashed file from disk, so the first boot after restoring a large volume is heavier than
usual.

## Rollback

The binary and the SPA ship in the same image, so rolling back the image rolls back both.
The database does not roll back — `AutoMigrate` has already added the new columns, and an
older binary simply ignores them. That is fine for additive changes and is the only reason
rollback is safe at all.
