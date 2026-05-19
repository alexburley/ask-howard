# Claude Code Instructions

## Architecture

Hexagonal (ports and adapters). Dependency flow: `adapter` → `port` ← `service` → `domain`.

```
internal/
├── domain/                         Pure Go types
├── port/inbound/                   Use case interfaces
├── port/outbound/                  Driven interfaces
├── service/                        Application services
└── adapter/
    ├── inbound/httpserver/         HTTP adapter (nickbryan/httputil)
    └── outbound/
        ├── postgres/               pgx/v5 repositories
        └── s3/                     Scaleway / MinIO adapter
```

## Testing

- **No stub-based unit tests for adapters.** Inbound and outbound adapters are tested with functional tests only.
- Functional tests live alongside the handler code they test (e.g. `handler/health_test.go`), tagged `//go:build functional`.
- Functional tests spin up real infrastructure via `testutil.NewPostgresContainer` (testcontainers-go).
- `make test` runs functional tests (`-tags functional`). `make test-unit` runs unit tests only (no Docker needed).
- Domain and service packages use plain unit tests with no build tag.
- **Always run `make test` (functional) before considering a feature complete.** Docker must be running.

## Linting

- Linter: `golangci-lint` v2 with config in `.golangci.yml`. Run with `make lint`.
- Auto-fix formatting with `make fmt` (runs `golangci-lint fmt`, which applies `gofumpt` and `goimports`).
- **Always run `make lint` before finishing a task.** Fix all issues; do not add `//nolint` directives without a clear reason.
- Enabled linters beyond the standard set: `bodyclose`, `contextcheck`, `err113`, `goconst`, `gocritic`, `misspell`, `nilerr`, `noctx`, `prealloc`, `revive`, `thelper`, `unconvert`, `unparam`, `wrapcheck`.
- `revive` doc-comment rules (`exported`, `package-comments`) are disabled — consistent with the no-comments convention.
- `err113`: errors must be defined as static `var Err… = errors.New(…)` or wrapped with `%w`; bare `fmt.Errorf("literal")` is not allowed outside `internal/testutil`.

## HTTP

- Router: `github.com/nickbryan/httputil`. Handlers use `httputil.NewHandler` with typed request/response generics.
- `httputil.RequestEmpty` = `Request[struct{}, struct{}]` for handlers with no body or params.
- `*http.Request` is embedded in the request type — use `r.Context()` directly.
- Error responses use `problem.DetailedError` (RFC 9457) from `github.com/nickbryan/httputil/problem`.
- Register endpoint groups with a prefix: `httputil.EndpointGroup(...).WithPrefix("/api")`.

## Conventions

- All constant string values in responses are uppercase (e.g. `"OK"`, not `"ok"`).
- Always update `README.md` when making structural, tooling, or architectural changes.

## Database & Migrations

- Schema and Atlas config live in `internal/adapter/outbound/postgres/` (`schema.hcl`, `atlas.hcl`).
- Migrations live in `internal/adapter/outbound/postgres/migrations/`.
- `make migrate-diff name=<description>` generates a new migration.
- `make migrate-apply` applies locally. Migrations applied automatically in functional tests.
- Atlas CE uses a positional name argument: `atlas migrate diff --env local "name"` (no `--name` flag).

## Dev Workflow

- `make start` — starts Docker infra (Postgres + MinIO), then launches air + Vite via hivemind.
- `//go:build dev` proxies frontend to Vite at `:5173`; `//go:build !dev` embeds `web/dist/`.
- Go binary served on `:8080`; Vite proxies `/api` to it during development.

## HTTP Request Files

- Manual test requests live in `http/*.http` — one file per handler group.
- Compatible with VS Code REST Client and JetBrains HTTP Client.
