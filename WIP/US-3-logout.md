# US-3: Logout

> As an authenticated user, I want to log out so that my session is ended.

## Acceptance criteria

- `POST /api/auth/logout` clears the JWT cookie (sets it expired)
- Returns `200` regardless of whether the user was authenticated

## What already exists

- `token.Issue` in `httpserver/token/token.go` — a companion `Clear` function lives here
- `handler.AuthEndpoints` — new endpoint added here
- No service call required — this is purely a cookie operation

## Implementation plan

### 1. Add `Clear` to the token package

`internal/adapter/inbound/httpserver/token/token.go`:

```go
func Clear(w http.ResponseWriter) {
    http.SetCookie(w, &http.Cookie{
        Name:     CookieName,
        Value:    "",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
        MaxAge:   -1,
        Path:     "/",
    })
}
```

`MaxAge: -1` instructs the browser to delete the cookie immediately.

### 2. Add logout handler

`internal/adapter/inbound/httpserver/handler/auth.go` — new endpoint in `AuthEndpoints`:

```go
{
    Method: http.MethodPost,
    Path:   "/auth/logout",
    Handler: httputil.NewHandler(func(r httputil.RequestEmpty) (*httputil.Response, error) {
        token.Clear(r.ResponseWriter)
        return httputil.OK(map[string]string{"status": "OK"})
    }),
},
```

No guard needed — the endpoint is intentionally public. Calling logout without a cookie is a no-op.

### 3. Functional tests

`handler/auth_test.go` — new test function `TestLogout_Functional`:

- Logout without a cookie → 200 (idempotent)
- Register, then logout → 200, cookie cleared (`MaxAge` = -1 or `Expires` in the past)

Cookie clearing check:

```go
cookie := cookieByName(resp, "token")
if cookie == nil || cookie.MaxAge != -1 {
    t.Error("token cookie was not cleared")
}
```

### 4. HTTP file

`http/auth.http` — append:

```
### Logout
POST http://localhost:8080/api/auth/logout
```
