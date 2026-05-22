# Ask Howard

## Architecture

Hexagonal (ports and adapters):

```
internal/
├── domain/                     Pure Go types — value objects and sentinel errors
├── port/
│   ├── inbound/                Use case interfaces (e.g. AuthService)
│   └── outbound/               Driven interfaces (UserRepository)
├── service/                    Application services — implement inbound ports
└── adapter/
    ├── inbound/httpserver/     HTTP adapter using nickbryan/httputil
    └── outbound/
        └── postgres/           PostgreSQL repositories (pgx/v5 + sqlc)
```

Dependency flow: `adapter` → `port` ← `service` → `domain`

The frontend is a React + TypeScript (Vite) SPA embedded into the Go binary via `go:embed`. In dev mode the Go server proxies to the Vite dev server instead.

## Requirements

- Docker
- Make

That's it. Go, Node, golangci-lint, Atlas, and sqlc are all provided by Docker — nothing needs to be installed locally.

## Development

After cloning, install the git hooks (runs lint and unit tests before each commit):

```bash
make hooks
```

Then start the stack:

```bash
make start
```

This builds and starts three containers — Postgres, the Go API (with `air` hot-reload), and the Vite dev server — and streams their logs in one terminal. Ctrl+C stops everything.

- The API waits for Postgres to pass its healthcheck before starting.
- Vite proxies `/api` requests to the Go server, so HMR and API calls work together without CORS configuration.
- Go source changes rebuild the binary via `air` inside the container; React changes are picked up by Vite HMR.

| Service  | URL                   |
|----------|-----------------------|
| API      | http://localhost:8080 |
| Frontend | http://localhost:5173 |

To start fresh (wipe volumes):

```bash
make clean-start
```

## Migrations

Schema is managed with [Atlas Community Edition](https://atlasgo.io/community-edition) (Apache 2.0). The desired schema is declared in `internal/adapter/outbound/postgres/schema.hcl` — Atlas diffs it against migration history to generate SQL.

Atlas CE runs inside the `ci` container — no local install needed.

**Changing the schema:**

1. Edit `internal/adapter/outbound/postgres/schema.hcl`
2. Generate a migration:
   ```bash
   make migrate-diff name=describe_your_change
   ```
3. Review the generated SQL in `internal/adapter/outbound/postgres/migrations/`
4. Apply locally:
   ```bash
   make migrate-apply
   ```

Check migration status at any time:

```bash
make migrate-status
```

Migrations are applied automatically in tests via `testutil.NewPostgresContainer`, which uses the Atlas Go SDK against the testcontainers Postgres instance.

## Code generation

SQL queries live in `internal/adapter/outbound/postgres/queries/`. [sqlc](https://sqlc.dev) compiles them into type-safe Go code in `internal/adapter/outbound/postgres/db/` — do not edit those files by hand.

After changing a `.sql` file, regenerate:

```bash
make generate
```

## Testing

### Backend

Tests are split into two categories by build tag:

| Command | Tag | What runs |
|---------|-----|-----------|
| `make test` | `functional` | All tests including functional |
| `make t` | `functional` | Same, with pretty output and coverage summary |
| `make coverage` | `functional` | Full per-function coverage breakdown |
| `make test-unit` | _(none)_ | Unit tests only (no Docker needed) |

Both run inside the `ci` container. **Functional tests** (`//go:build functional`) spin up real infrastructure via [testcontainers-go](https://testcontainers.com/guides/getting-started-with-testcontainers-for-go/) — they use the Docker socket, not the compose Postgres. Migrations are applied automatically via `internal/testutil.NewPostgresContainer`.

### Frontend (E2E)

```bash
make e2e
```

Runs Playwright tests against the full running stack (requires `make start` first). The `playwright` Docker service uses `mcr.microsoft.com/playwright:v1.60.0-noble` — no local browser install needed.

E2E tests live in `web/e2e/`.

## Production build

```bash
make build
```

Builds the production Docker image tagged `ask-howard:local`. The multi-stage `Dockerfile` compiles the frontend, embeds it into the Go binary, and produces a minimal Alpine runtime image.

## Local infrastructure

`docker-compose.yml` provides:

| Service    | Port | Notes                                        |
|------------|------|----------------------------------------------|
| Postgres   | 5432 | ask-howard / ask-howard / ask-howard         |
| API        | 8080 | Go binary via `air` hot-reload (`-tags dev`) |
| Web        | 5173 | Vite dev server with HMR                     |
| MinIO      | 9000 / 9001 | S3-compatible object store / console   |

> **Adding npm packages:** the web container uses a named Docker volume for `node_modules`. After running `npm install` on the host, reset the volume so the container picks up the new packages:
> ```bash
> docker compose rm -f web && docker volume rm ask-howard_web_node_modules && docker compose up -d web
> ```
| Playwright | —    | Profile `e2e` only — started by `make e2e`  |

## Key dependencies

| Package | Purpose |
|---------|---------|
| [nickbryan/httputil](https://github.com/nickbryan/httputil) | Type-safe HTTP handlers, RFC 9457 error responses |
| [jackc/pgx/v5](https://github.com/jackc/pgx) | PostgreSQL driver and connection pool |
| [sqlc](https://sqlc.dev) | Type-safe SQL query generation |
| [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) | JWT signing and verification (HS256) |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) | bcrypt password hashing |
| [testcontainers-go](https://github.com/testcontainers/testcontainers-go) | Real Postgres containers for functional tests |
| React 19 + Vite 6 | Frontend SPA |
| [@playwright/test](https://playwright.dev) | E2E browser tests |
