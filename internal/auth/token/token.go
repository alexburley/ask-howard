package token

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alexburley/ask-howard/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

const (
	CookieName = "token"
	ttl        = 7 * 24 * time.Hour
)

func Issue(w http.ResponseWriter, secret auth.JWTSecret, userID string) error {
	now := time.Now()
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
		"iss": "ask-howard",
	}).SignedString(secret.Bytes())
	if err != nil {
		return fmt.Errorf("sign token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    signed,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(ttl.Seconds()),
		Path:     "/",
	})

	return nil
}
