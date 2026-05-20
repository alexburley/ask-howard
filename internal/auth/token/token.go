package token

import (
	"errors"
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

var (
	errUnexpectedSigningMethod = errors.New("unexpected signing method")
	errMissingSubjectClaim     = errors.New("missing subject claim")
)

func Parse(r *http.Request, secret auth.JWTSecret) (string, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return "", fmt.Errorf("read cookie: %w", err)
	}

	t, err := jwt.ParseWithClaims(cookie.Value, &jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errUnexpectedSigningMethod
		}
		return secret.Bytes(), nil
	})
	if err != nil || !t.Valid {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	sub, err := t.Claims.GetSubject()
	if err != nil || sub == "" {
		return "", errMissingSubjectClaim
	}

	return sub, nil
}

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
