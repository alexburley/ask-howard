package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alexburley/ask-howard/internal/auth"
	"github.com/alexburley/ask-howard/internal/auth/token"
	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/inbound"
	"github.com/nickbryan/httputil"
	"github.com/nickbryan/httputil/problem"
)

type registerBody struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func AuthEndpoints(svc inbound.AuthService, jwtSecret auth.JWTSecret) []httputil.Endpoint {
	return []httputil.Endpoint{
		{
			Method: http.MethodPost,
			Path:   "/auth/register",
			Handler: httputil.NewHandler(func(r httputil.RequestData[registerBody]) (*httputil.Response, error) {
				user, err := svc.Register(r.Context(), r.Data.Email, r.Data.Password)
				if err != nil {
					switch {
					case errors.Is(err, domain.ErrEmailTaken):
						return nil, (&problem.DetailedError{
							Type:   "https://ask-howard.io/problems/email-taken",
							Title:  "Email Already Registered",
							Status: http.StatusConflict,
						}).WithDetail("An account with this email address already exists")
					case errors.Is(err, domain.ErrInvalidEmail):
						return nil, (&problem.DetailedError{
							Type:   "https://ask-howard.io/problems/invalid-input",
							Title:  "Invalid Input",
							Status: http.StatusUnprocessableEntity,
						}).WithDetail("Invalid email address")
					case errors.Is(err, domain.ErrPasswordTooShort):
						return nil, (&problem.DetailedError{
							Type:   "https://ask-howard.io/problems/invalid-input",
							Title:  "Invalid Input",
							Status: http.StatusUnprocessableEntity,
						}).WithDetail("Password must be at least 8 characters")
					}
					return nil, fmt.Errorf("register user: %w", err)
				}

				if err := token.Issue(r.ResponseWriter, jwtSecret, user.ID.String()); err != nil {
					return nil, fmt.Errorf("issue token: %w", err)
				}

				return httputil.Created(userResponse{ID: user.ID.String(), Email: user.Email.String()})
			}),
		},
	}
}
