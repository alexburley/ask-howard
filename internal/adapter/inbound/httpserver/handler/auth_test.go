//go:build functional

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver"
	"github.com/alexburley/ask-howard/internal/adapter/outbound/postgres"
	"github.com/alexburley/ask-howard/internal/service"
	"github.com/alexburley/ask-howard/internal/testutil"
)

const testJWTSecret = "test-secret-at-least-32-chars-long"

func TestRegister_Functional(t *testing.T) {
	pg := testutil.NewPostgresContainer(t)

	pool, err := postgres.NewPool(context.Background(), pg.ConnectionString)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	authSvc := service.NewAuthService(postgres.NewUserRepository(pool))
	srv := httpserver.NewServer(slog.New(slog.NewTextHandler(io.Discard, nil)), pool, authSvc, testJWTSecret)
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)

	t.Run("created with valid credentials", func(t *testing.T) {
		resp := postRegister(t, ts, "alice@example.com", "password123")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
		}

		var body map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["email"] != "alice@example.com" {
			t.Errorf("email = %q, want %q", body["email"], "alice@example.com")
		}
		if body["id"] == "" {
			t.Error("id is empty")
		}

		cookie := cookieByName(resp, "token")
		if cookie == nil {
			t.Fatal("token cookie not set")
		}
		if !cookie.HttpOnly {
			t.Error("token cookie is not HttpOnly")
		}
	})

	t.Run("conflict on duplicate email", func(t *testing.T) {
		postRegister(t, ts, "bob@example.com", "password123")
		resp := postRegister(t, ts, "bob@example.com", "password123")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusConflict)
		}
	})

	t.Run("unprocessable on short password", func(t *testing.T) {
		resp := postRegister(t, ts, "carol@example.com", "short")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

	t.Run("unprocessable on invalid email", func(t *testing.T) {
		resp := postRegister(t, ts, "not-an-email", "password123")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})
}

func postRegister(t *testing.T, ts *httptest.Server, email, password string) *http.Response {
	t.Helper()

	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	resp, err := ts.Client().Post(ts.URL+"/api/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/auth/register: %v", err)
	}
	return resp
}

func cookieByName(resp *http.Response, name string) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}
