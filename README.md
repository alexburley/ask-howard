# Ask Howard

## Architecture

Hexagonal (ports and adapters):

```
internal/
├── domain/                     Pure Go types — Document, Person, Claim, etc.
├── port/
│   ├── inbound/                Use case interfaces (e.g. DocumentService)
│   └── outbound/               Driven interfaces (DocumentRepository, ObjectStore)
├── service/                    Application services — implement inbound ports
└── adapter/
    ├── inbound/httpserver/     HTTP adapter using nickbryan/httputil
    └── outbound/
        ├── postgres/           PostgreSQL repositories (pgx/v5)
        └── s3/                 Scaleway Object Storage adapter
```

Dependency flow: `adapter` → `port` ← `service` → `domain`

The frontend is a React + TypeScript (Vite) SPA embedded into the Go binary via `go:embed`. In dev mode the Go server proxies to the Vite dev server instead.

## Requirements

- Go 1.26+ (`/opt/homebrew/bin/go` if installed via Homebrew)
- Node 22+
- Docker (for local infrastructure and tests)

## Development

Add Go 1.26 to your shell if not already set:

```bash
echo 'export PATH=/opt/homebrew/bin:$PATH' >> ~/.zshrc && source ~/.zshrc
```

Install tooling and dependencies:

```bash
go install github.com/air-verse/air@latest   # Go hot reload
brew install hivemind                         # process manager (no tmux required)
go mod download
make web-install
```

Then just:

```bash
make start
```

This starts Postgres and MinIO (waiting for healthchecks), then launches the Go API with hot reload and the Vite dev server together in one terminal. Ctrl+C stops everything.

The Vite dev server proxies `/api` requests to the Go server on `:8080`, so HMR and API calls work together without CORS configuration.

## Migrations

Schema is managed with [Atlas Community Edition](https://atlasgo.io/community-edition) (Apache 2.0). The desired schema is declared in `schema.hcl` — Atlas diffs it against migration history to generate SQL.

Install Atlas CE if you haven't already:

```bash
curl -sSf https://atlasgo.sh | sh -s -- --community
# move the binary to ~/go/bin or anywhere on your PATH
```

**Changing the schema:**

1. Edit `schema.hcl`
2. Generate a migration:
   ```bash
   make migrate-diff name=describe_your_change
   ```
3. Review the generated SQL in `migrations/`
4. Apply locally:
   ```bash
   make migrate-apply
   ```

Check migration status at any time:

```bash
make migrate-status
```

Migrations are applied automatically in tests via `testutil.NewPostgresContainer`, which uses the Atlas Go SDK against the testcontainers Postgres instance.

## Testing

Tests are split into two categories by build tag:

| Command | Tag | What runs |
|---------|-----|-----------|
| `make test` | `functional` | All tests including functional (requires Docker) |
| `make test-unit` | _(none)_ | Unit tests only, no containers needed |

**Functional tests** (`//go:build functional`) spin up real infrastructure via [testcontainers-go](https://testcontainers.com/guides/getting-started-with-testcontainers-for-go/). Docker must be running. Migrations are applied automatically before each test via `internal/testutil.NewPostgresContainer`.

**Unit tests** use stubs and `httptest` only — no Docker required.

## Production build

```bash
make build
```

This runs `npm run build` inside `web/`, then compiles the Go binary with the frontend assets embedded. The result is a single self-contained binary at `bin/ask-howard`.

## Docker

```bash
make docker-build        # builds image tagged ask-howard:local
docker run -p 8080:8080 ask-howard:local
```

The multi-stage `Dockerfile` builds the frontend, then the Go binary, and produces a minimal Alpine runtime image.

## Local infrastructure

`docker-compose.yml` provides:

| Service  | Port | Credentials                              |
|----------|------|------------------------------------------|
| Postgres | 5432 | ask-howard / ask-howard / ask-howard     |

## Key dependencies

| Package | Purpose |
|---------|---------|
| [nickbryan/httputil](https://github.com/nickbryan/httputil) | Type-safe HTTP handlers, RFC 9457 error responses |
| [jackc/pgx/v5](https://github.com/jackc/pgx) | PostgreSQL driver and connection pool |
| [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) | JWT signing and verification (HS256) |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) | bcrypt password hashing |
| [testcontainers-go](https://github.com/testcontainers/testcontainers-go) | Real Postgres containers for API tests |
| React 19 + Vite 6 | Frontend SPA |
