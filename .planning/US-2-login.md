# US-2: Login

> As a registered user, I want to log in with my email and password so that I can authenticate my session.

## Acceptance criteria

- `POST /api/auth/login` accepts `{"email": "...", "password": "..."}`
- Returns `200` with an HttpOnly JWT cookie and `{"id": "...", "email": "..."}` on success
- Returns `401` for invalid credentials — same message for unknown email or wrong password (no enumeration)

## What already exists

- `domain.User`, `domain.ErrEmailTaken`
- `port/outbound.UserRepository` interface — `FindByEmail` stubbed in postgres adapter
- `port/inbound.AuthService` interface — needs `Login` added
- `service.AuthService` — needs `Login` implemented
- `token.Issue` — reused as-is
- `handler.AuthEndpoints` — new endpoint added here

## Implementation plan

### 1. Add sentinel error

`internal/domain/user.go` — add alongside `ErrEmailTaken`:

```go
var ErrInvalidCredentials = errors.New("invalid email or password")
```

### 2. Extend the inbound port

`internal/port/inbound/auth.go` — add `Login` to the `AuthService` interface:

```go
Login(ctx context.Context, email, password string) (domain.User, error)
```

### 3. Implement `FindByEmail` in the postgres adapter

`internal/adapter/outbound/postgres/user.go` — replace the `panic`:

```go
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
    var user domain.User
    err := r.db.QueryRow(ctx,
        `SELECT id, email, password_hash FROM users WHERE email = $1`,
        email,
    ).Scan(&user.ID, &user.Email, &user.PasswordHash)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return domain.User{}, domain.ErrInvalidCredentials
        }
        return domain.User{}, fmt.Errorf("find user by email: %w", err)
    }
    return user, nil
}
```

### 4. Implement `Login` in the service

`internal/service/auth.go`:

```go
func (s *AuthService) Login(ctx context.Context, email, password string) (domain.User, error) {
    user, err := s.users.FindByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, domain.ErrInvalidCredentials) {
            return domain.User{}, domain.ErrInvalidCredentials
        }
        return domain.User{}, fmt.Errorf("find user: %w", err)
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return domain.User{}, domain.ErrInvalidCredentials
    }

    return user, nil
}
```

Note: wrong password returns `ErrInvalidCredentials`, not a bcrypt error, to prevent enumeration.

### 5. Add login handler

`internal/adapter/inbound/httpserver/handler/auth.go` — new endpoint in `AuthEndpoints`:

```go
{
    Method: http.MethodPost,
    Path:   "/auth/login",
    Handler: httputil.NewHandler(func(r httputil.RequestData[loginBody]) (*httputil.Response, error) {
        user, err := svc.Login(r.Context(), r.Data.Email, r.Data.Password)
        if err != nil {
            if errors.Is(err, domain.ErrInvalidCredentials) {
                return nil, &problem.DetailedError{
                    Type:   "https://ask-howard.io/problems/invalid-credentials",
                    Title:  "Invalid Credentials",
                    Status: http.StatusUnauthorized,
                }
            }
            return nil, err
        }

        if err := token.Issue(r.ResponseWriter, jwtSecret, user.ID); err != nil {
            return nil, err
        }

        return httputil.OK(userResponse{ID: user.ID, Email: user.Email})
    }),
},
```

`loginBody` reuses the same shape as `registerBody` — extract to a shared `authBody` struct.

### 6. Functional tests

`handler/auth_test.go` — new test function `TestLogin_Functional`:

- Happy path: register then login → 200, cookie set, correct body
- Wrong password → 401
- Unknown email → 401
- Missing fields → 422

### 7. HTTP file

`http/auth.http` — append:

```
### Login
POST http://localhost:8080/api/auth/login
Content-Type: application/json

{"email": "user@example.com", "password": "password123"}

### Login - wrong password (expect 401)
POST http://localhost:8080/api/auth/login
Content-Type: application/json

{"email": "user@example.com", "password": "wrongpassword"}
```
