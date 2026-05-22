package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/nickbryan/httputil/problem"
)

type contextKey struct{}

const (
	ProblemUnauthorizedType  = "https://ask-howard.io/problems/unauthorized"
	ProblemUnauthorizedTitle = "Unauthorized"
)

func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(contextKey{}).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, &problem.DetailedError{
			Type:   ProblemUnauthorizedType,
			Title:  ProblemUnauthorizedTitle,
			Status: http.StatusUnauthorized,
		}
	}
	return id, nil
}
