# US-4: Current User (Me)

> As an authenticated user, I want the frontend to know who I am on page load without re-authenticating.

## Acceptance criteria

- `GET /api/auth/me` returns `{"id": "...", "email": "..."}` if the JWT cookie is valid
- Returns `401` if the cookie is missing or invalid

## What already exists

- `token.Issue` in `httpserver/token/token.go` — a companion `Parse` function lives here
- `port/outbound.UserRepository` — `FindByID` stubbed in postgres adapter
- `port/inbound.AuthService` — needs `GetByID` added
- `service.AuthService` — needs `GetByID` implemented
- `handler.AuthEndpoints` — new endpoint added here

## Implementation plan

### 1. Add `Parse` to the token package

`internal/adapter/inbound/httpserver/token/token.go`:

```go
// Parse validates the token cookie and returns the subject (user ID).
// Returns an error if the cookie is missing, the token is invalid, or it has expired.
func Parse(r *http.Request, secret string) (string, error) {
    cookie, err := r.Cookie(CookieName)
    if err != nil {
        return "", fmt.Errorf("read cookie: %w", err)
    }

    t, err := jwt.ParseWithClaims(cookie.Value, &jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return []byte(secret), nil
    })
    if err != nil || !t.Valid {
        return "", fmt.Errorf("invalid token: %w", err)
    }

    sub, err := t.Claims.GetSubject()
    if err != nil || sub == "" {
        return "", fmt.Errorf("missing subject claim")
    }

    return sub, nil
}
```

### 2. Add sentinel error

`internal/domain/user.go` — add alongside existing errors:

```go
var ErrUserNotFound = errors.New("user not found")
```

### 3. Implement `FindByID` in the postgres adapter

`internal/adapter/outbound/postgres/user.go` — replace the `panic`:

```go
func (r *UserRepository) FindByID(ctx context.Context, id string) (domain.User, error) {
    var user domain.User
    err := r.db.QueryRow(ctx,
        `SELECT id, email, password_hash FROM users WHERE id = $1`,
        id,
    ).Scan(&user.ID, &user.Email, &user.PasswordHash)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return domain.User{}, domain.ErrUserNotFound
        }
        return domain.User{}, fmt.Errorf("find user by id: %w", err)
    }
    return user, nil
}
```

### 4. Extend the inbound port

`internal/port/inbound/auth.go` — add `GetByID` to the `AuthService` interface:

```go
GetByID(ctx context.Context, id string) (domain.User, error)
```

### 5. Implement `GetByID` in the service

`internal/service/auth.go`:

```go
func (s *AuthService) GetByID(ctx context.Context, id string) (domain.User, error) {
    user, err := s.users.FindByID(ctx, id)
    if err != nil {
        return domain.User{}, fmt.Errorf("find user: %w", err)
    }
    return user, nil
}
```

### 6. Add me handler

`internal/adapter/inbound/httpserver/handler/auth.go` — new endpoint in `AuthEndpoints`:

```go
{
    Method: http.MethodGet,
    Path:   "/auth/me",
    Handler: httputil.NewHandler(func(r httputil.RequestEmpty) (*httputil.Response, error) {
        userID, err := token.Parse(r.Request, jwtSecret)
        if err != nil {
            return nil, &problem.DetailedError{
                Type:   "https://pulse.app/problems/unauthorized",
                Title:  "Unauthorized",
                Status: http.StatusUnauthorized,
            }
        }

        user, err := svc.GetByID(r.Context(), userID)
        if err != nil {
            if errors.Is(err, domain.ErrUserNotFound) {
                return nil, &problem.DetailedError{
                    Type:   "https://pulse.app/problems/unauthorized",
                    Title:  "Unauthorized",
                    Status: http.StatusUnauthorized,
                }
            }
            return nil, err
        }

        return httputil.OK(userResponse{ID: user.ID, Email: user.Email})
    }),
},
```

Note: `ErrUserNotFound` also returns 401 (not 404) — the token refers to a deleted account, which is an authentication failure from the client's perspective.

### 7. Functional tests

`handler/auth_test.go` — new test function `TestMe_Functional`:

- Register → GET /me with cookie → 200, correct body
- GET /me with no cookie → 401
- GET /me with tampered token → 401

Cookie replay in tests:

```go
req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/auth/me", nil)
req.AddCookie(cookieByName(registerResp, "token"))
resp, _ := ts.Client().Do(req)
```

### 8. HTTP file

`http/auth.http` — append:

```
### Me (requires token cookie from login/register)
GET http://localhost:8080/api/auth/me
```

---

## Note on auth middleware

The guard/middleware pattern (for protecting non-auth routes) is deferred until the first protected route is added. When that time comes:

- Add `token.Parse` call inside an `httputil.GuardFunc`
- Inject the user ID into `context.Context` via a typed key
- Apply with `.WithGuard(authGuard)` on the protected `EndpointGroup`
- The `me` handler above can then be refactored to read from context rather than parsing the token directly
