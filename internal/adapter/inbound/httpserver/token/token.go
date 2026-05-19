package token

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	CookieName = "token"
	ttl        = 7 * 24 * time.Hour
)

func Issue(w http.ResponseWriter, secret, userID string) error {
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(ttl).Unix(),
	}).SignedString([]byte(secret))
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
