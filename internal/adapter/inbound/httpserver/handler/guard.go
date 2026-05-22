package handler

import (
	"net/http"

	"github.com/alexburley/ask-howard/internal/auth"
	"github.com/alexburley/ask-howard/internal/auth/token"
	"github.com/google/uuid"
	"github.com/nickbryan/httputil"
	"github.com/nickbryan/httputil/problem"
)

// NewAuthGuard returns a Guard that validates the JWT cookie and injects the
// user ID into the request context. Endpoints protected by this guard can read
// the user ID via auth.UserIDFromContext.
func NewAuthGuard(jwtSecret auth.JWTSecret) httputil.GuardFunc {
	return func(r *http.Request) (*http.Request, error) {
		rawID, err := token.Parse(r, jwtSecret)
		if err != nil {
			return nil, &problem.DetailedError{
				Type:   auth.ProblemUnauthorizedType,
				Title:  auth.ProblemUnauthorizedTitle,
				Status: http.StatusUnauthorized,
			}
		}

		userID, err := uuid.Parse(rawID)
		if err != nil {
			return nil, &problem.DetailedError{
				Type:   auth.ProblemUnauthorizedType,
				Title:  auth.ProblemUnauthorizedTitle,
				Status: http.StatusUnauthorized,
			}
		}

		return r.WithContext(auth.WithUserID(r.Context(), userID)), nil
	}
}
