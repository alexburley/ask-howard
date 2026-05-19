# Codebase Review

> Reviewed against: Go best practices, DDD/hexagonal architecture, and developer experience.
> Status as of: 2026-05-19

The overall structure is sound — the hexagonal layout is correctly implemented, the testing strategy is well-considered, and security defaults are good. The items below are ordered roughly by impact.

---

## 1. Domain model leaks infrastructure concerns

**`internal/domain/user.go` and `internal/port/outbound/user.go`**

`UserRepository.Create` takes `(ctx, email, passwordHash string)`. The password hash is an infrastructure artefact — it belongs to the persistence layer, not the domain contract. The outbound port is the seam between your service and the database; it should speak in domain terms.

**Suggested change:**

```go
// port/outbound/user.go
type UserRepository interface {
    Create(ctx context.Context, user domain.User) (domain.User, error)
    FindByEmail(ctx context.Context, email string) (domain.User, error)
    FindByID(ctx context.Context, id string) (domain.User, error)
}
```

The service builds the `domain.User` (with hashed password), then hands it to the repository. The repository decides how to persist it. This removes the `passwordHash` primitive from the port signature and keeps `Create` aligned with domain language.

---

## 2. `User.ID` typed as `string` — should reflect the actual type

**`internal/domain/user.go`**

The schema uses UUID (`gen_random_uuid()`), but the domain type uses `string`. This means the service and handler layers have no compile-time guarantee that an ID is a valid UUID — any string passes. At minimum, use `github.com/google/uuid.UUID` (already an indirect dependency):

```go
import "github.com/google/uuid"

type User struct {
    ID           uuid.UUID
    Email        string
    PasswordHash string
}
```

This also surfaces the `uuid` package as a direct dependency in `go.mod`, which correctly reflects its use.

---

## 3. Value objects: `Email` and `Password` carry no invariants

**`internal/domain/`**

`email` and `password` are bare strings throughout the call chain. Validation currently lives only in the HTTP layer (`validate:"required,email"`, `validate:"min=8"`). If you ever add a CLI, a gRPC adapter, or a background job that creates users, those invariants are silently absent.

Consider simple value objects in the domain:

```go
type Email struct{ value string }

func NewEmail(s string) (Email, error) {
    if !isValidEmail(s) {
        return Email{}, ErrInvalidEmail
    }
    return Email{value: s}, nil
}

func (e Email) String() string { return e.value }
```

The HTTP handler still validates eagerly at the boundary (fast failure, good UX), but the domain types enforce invariants unconditionally. The service constructs them, and a failure there means something bypassed the HTTP layer.

---

## 4. Unimplemented repository methods panic in production code

**`internal/adapter/outbound/postgres/user.go:44-49`**

`FindByEmail` and `FindByID` both `panic("not implemented")`. These compile into the production binary and will crash the server the moment any code path reaches them. Even in WIP, prefer returning a sentinel error:

```go
var ErrNotImplemented = errors.New("not implemented")

func (r *UserRepository) FindByEmail(_ context.Context, _ string) (domain.User, error) {
    return domain.User{}, ErrNotImplemented
}
```

This lets tests fail descriptively rather than crashing the process.

---

## 5. `token` package is mislocated

**`internal/adapter/inbound/httpserver/token/`**

The token package currently lives inside the HTTP adapter. The login handler (US-2), logout, and any auth middleware will all need it. Once you implement authentication middleware, that middleware will live at the `httpserver` level or higher — and importing a sub-package of an adapter from its own parent is awkward.

Move it to `internal/adapter/inbound/httpserver/` as a sibling of `handler/`, or extract it to `internal/auth/token/` if you anticipate it being shared across adapters (e.g., a future gRPC adapter).

---

## 6. `jwtSecret` is passed as a raw `string` through three layers

**`cmd/server/main.go → httpserver.NewServer → handler.AuthEndpoints`**

A bare `string` for a secret gives no type safety and invites accidental logging or serialisation. A small wrapper type prevents misuse and makes the dependency explicit:

```go
// internal/auth/secret.go (or alongside the token package)
type JWTSecret struct{ value string }

func NewJWTSecret(s string) JWTSecret { return JWTSecret{value: s} }
func (s JWTSecret) String() string    { return "[REDACTED]" } // safe logging
func (s JWTSecret) Bytes() []byte     { return []byte(s.value) }
```

---

## 7. Configuration is scattered across `main.go`

**`cmd/server/main.go`**

Config is assembled inline with `os.Getenv` calls interleaved with construction logic. As the app grows (port, log level, TLS, etc.), this becomes hard to validate and test. A dedicated config struct loaded at startup is idiomatic Go:

```go
type Config struct {
    DatabaseURL string
    JWTSecret   string
    Port        string
}

func loadConfig() Config {
    c := Config{
        DatabaseURL: env("DATABASE_URL", "postgres://ask-howard:ask-howard@localhost:5432/ask-howard?sslmode=disable"),
        JWTSecret:   env("JWT_SECRET", ""),
        Port:        env("PORT", "8080"),
    }
    if c.JWTSecret == "" {
        c.JWTSecret = "dev-secret-do-not-use-in-production"
        // warn
    }
    return c
}
```

This also makes the config unit-testable independently of the running server, and gives you one place to add validation (e.g. require `JWT_SECRET` in production).

---

## 8. `golang-jwt/jwt` should be a direct dependency

**`go.mod:81`**

`golang-jwt/jwt/v5` is listed as `// indirect` but is used directly in `internal/adapter/inbound/httpserver/token/token.go`. Run `go mod tidy` — it should become a direct dependency. Indirect means nothing in the module actually imports it directly, which is misleading here.

---

## 9. Tests spin up a Postgres container per test function

**`internal/adapter/inbound/httpserver/handler/auth_test.go`**

Each top-level `TestXxx_Functional` function calls `testutil.NewPostgresContainer(t)`, which starts a new container. For two test functions in the same file, that's two containers sequentially. As the test suite grows this compounds.

Consider a `TestMain` approach with a package-level container, or use `t.Parallel()` combined with a shared container seeded per test via schema reset or subtransactions. At minimum, the current tests should call `t.Parallel()` so the Go test runner can schedule them concurrently.

Also: in `TestRegister_Functional`, the "conflict on duplicate email" subtest calls `postRegister` for `bob@example.com` a first time without asserting it succeeded (201). If that first registration fails silently, the test proves nothing. Assert the first call returns 201.

---

## 10. `HealthChecker` interface defined in the wrong layer

**`internal/adapter/inbound/httpserver/server.go:16`**

The `HealthChecker` interface is defined inside the `httpserver` adapter. Interfaces in Go should be owned by the consumer (which is correct here), but placing an outbound-style interface inside an inbound adapter is confusing. Since this is a driven port (the server drives the DB ping), define it in `internal/port/outbound/` alongside `UserRepository`, or — given its simplicity — keep it in the httpserver package but in a dedicated `ports.go` file rather than `server.go`.

---

## 11. JWT claims missing `iat` and `iss`

**`internal/adapter/inbound/httpserver/token/token.go:17`**

The token only carries `sub` and `exp`. Standard practice includes:
- `iat` (issued at) — enables detection of tokens issued before a password change
- `iss` (issuer) — defensive against token confusion if you ever run multiple services

```go
jwt.MapClaims{
    "sub": userID,
    "iat": time.Now().Unix(),
    "exp": time.Now().Add(ttl).Unix(),
    "iss": "ask-howard",
}
```

---

## 12. No CI pipeline

There is no `.github/workflows/` or equivalent. Given the project has `make test` (Docker), `make lint`, and `make build`, a basic GitHub Actions workflow that runs all three on push/PR would catch regressions early and serves as living documentation of what "passing" means.

Minimum viable workflow:
1. `golangci-lint` (no Docker needed)
2. `go test ./...` (unit tests, no Docker)
3. `go test -tags functional ./...` (functional tests, use `services: postgres` in the workflow)

---

## 13. No `.env.example` file

**Project root**

`DATABASE_URL` and `JWT_SECRET` are documented only implicitly in `main.go`. A `.env.example` (committed, secrets-free) makes onboarding explicit and pairs well with `direnv` or `dotenv` for local development.

---

## 14. `WIP/` directory — consider a different convention

**`WIP/US-2-login.md`, etc.**

The WIP files are well-written implementation plans. However, checked-in markdown spec files in a `WIP/` directory tend to drift out of sync with the code and clutter the tree. The content is valuable; the location may not scale. Options:
- Move to GitHub Issues or a project board (links back to the code)
- Keep them but move into a `docs/` or `.planning/` directory (ignored by most tooling)
- Delete after implementation is complete (the plan's value is in guiding the work, not in archiving it)

---

## Summary

| # | Area | Effort | Impact |
|---|------|--------|--------|
| 1 | Remove `passwordHash` from outbound port | Small | Architecture clarity |
| 2 | `User.ID` as `uuid.UUID` | Small | Type safety |
| 3 | Value objects for `Email`/`Password` | Medium | Domain integrity |
| 4 | Replace panics with sentinel errors | Trivial | Safety |
| 5 | Relocate `token` package | Small | Maintainability |
| 6 | Typed `JWTSecret` | Small | Safety |
| 7 | Config struct in `main.go` | Small | Readability |
| 8 | Fix indirect JWT dependency | Trivial | Hygiene |
| 9 | Shared container / parallel tests | Medium | Test speed |
| 10 | `HealthChecker` placement | Trivial | Clarity |
| 11 | JWT `iat`/`iss` claims | Trivial | Security posture |
| 12 | CI pipeline | Medium | Confidence |
| 13 | `.env.example` | Trivial | DX |
| 14 | `WIP/` convention | Small | DX |
