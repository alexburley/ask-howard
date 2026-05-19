package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver/token"
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

func AuthEndpoints(svc inbound.AuthService, jwtSecret string) []httputil.Endpoint {
	return []httputil.Endpoint{
		{
			Method: http.MethodPost,
			Path:   "/auth/register",
			Handler: httputil.NewHandler(func(r httputil.RequestData[registerBody]) (*httputil.Response, error) {
				user, err := svc.Register(r.Context(), r.Data.Email, r.Data.Password)
				if err != nil {
					if errors.Is(err, domain.ErrEmailTaken) {
						return nil, (&problem.DetailedError{
							Type:   "https://ask-howard.io/problems/email-taken",
							Title:  "Email Already Registered",
							Status: http.StatusConflict,
						}).WithDetail("An account with this email address already exists")
					}
					return nil, fmt.Errorf("register user: %w", err)
				}

				if err := token.Issue(r.ResponseWriter, jwtSecret, user.ID); err != nil {
					return nil, fmt.Errorf("issue token: %w", err)
				}

				return httputil.Created(userResponse{ID: user.ID, Email: user.Email})
			}),
		},
	}
}
