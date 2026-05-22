package handler

import (
	"net/http"

	"github.com/alexburley/ask-howard/internal/auth"
	"github.com/alexburley/ask-howard/internal/auth/token"
	"github.com/google/uuid"
	"github.com/nickbryan/httputil/problem"
)

func currentUserID(r *http.Request, jwtSecret auth.JWTSecret) (uuid.UUID, error) {
	rawID, err := token.Parse(r, jwtSecret)
	if err != nil {
		return uuid.UUID{}, &problem.DetailedError{
			Type:   problemUnauthorizedType,
			Title:  problemUnauthorizedTitle,
			Status: http.StatusUnauthorized,
		}
	}

	userID, err := uuid.Parse(rawID)
	if err != nil {
		return uuid.UUID{}, &problem.DetailedError{
			Type:   problemUnauthorizedType,
			Title:  problemUnauthorizedTitle,
			Status: http.StatusUnauthorized,
		}
	}

	return userID, nil
}
