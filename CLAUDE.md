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
- **Always run `make test` (functional) and `make lint` before considering a feature complete.** 


## HTTP

- Router: `github.com/nickbryan/httputil`. Handlers use `httputil.NewHandler` with typed request/response generics.
- `httputil.RequestEmpty` = `Request[struct{}, struct{}]` for handlers with no body or params.
- `*http.Request` is embedded in the request type — use `r.Context()` directly.
- Error responses use `problem.DetailedError` (RFC 9457) from `github.com/nickbryan/httputil/problem`.

## Conventions

- All constant string values in responses are uppercase (e.g. `"OK"`, not `"ok"`).
- Always update `README.md` when making structural, tooling, or architectural changes.
- **Always prompt to commit after completing a task.** Remind the user to commit if they haven't.

## Code Generation

- sqlc generates type-safe query code from `internal/adapter/outbound/postgres/queries/*.sql`.
- Generated files live in `internal/adapter/outbound/postgres/db/` — do not edit by hand.
- Run `make generate` (or `make sqlc`) after changing any `.sql` query file.
- `make generate` is the single entry point for all code generation tools.

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

## Frontend

- Stack: React 19, TypeScript, Vite. No client-side router — use conditional rendering.
- Source layout: `web/src/auth/` (context + API client), `web/src/components/`, `web/src/hooks/`, `web/src/pages/`.
- Auth state is managed by `AuthProvider` (wraps the app in `main.tsx`). Use `useAuth()` to access `user`, `isLoading`, `login`, `register`, `logout`.
- API calls go through `web/src/auth/api.ts` — throw `AuthError` with a typed `code` field; never expose raw fetch errors to components.
- E2E tests use Playwright and live in `web/e2e/`. Run with `make e2e` (requires `make start` running first).
- Each e2e test that creates a user must use a unique email (`test-${Date.now()}@example.com`) — there is no test DB reset between runs.
- TypeScript must compile cleanly (`npx tsc --noEmit`) before considering frontend work done.

## HTTP Request Files

- Manual test requests live in `http/*.http` — one file per handler group.
