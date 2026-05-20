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

type authBody struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type loginBody struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"`
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
			Handler: httputil.NewHandler(func(r httputil.RequestData[authBody]) (*httputil.Response, error) {
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
		{
			Method: http.MethodPost,
			Path:   "/auth/login",
			Handler: httputil.NewHandler(func(r httputil.RequestData[loginBody]) (*httputil.Response, error) {
				user, err := svc.Login(r.Context(), r.Data.Email, r.Data.Password)
				if err != nil {
					if errors.Is(err, domain.ErrInvalidCredentials) {
						return nil, (&problem.DetailedError{
							Type:   "https://ask-howard.io/problems/invalid-credentials",
							Title:  "Invalid Credentials",
							Status: http.StatusUnauthorized,
						}).WithDetail("Invalid email or password")
					}
					return nil, fmt.Errorf("login: %w", err)
				}

				if err := token.Issue(r.ResponseWriter, jwtSecret, user.ID.String()); err != nil {
					return nil, fmt.Errorf("issue token: %w", err)
				}

				return httputil.OK(userResponse{ID: user.ID.String(), Email: user.Email.String()})
			}),
		},
		{
			Method: http.MethodPost,
			Path:   "/auth/logout",
			Handler: httputil.NewHandler(func(r httputil.RequestEmpty) (*httputil.Response, error) {
				token.Clear(r.ResponseWriter)
				return httputil.OK(map[string]string{"status": "OK"})
			}),
		},
		{
			Method: http.MethodGet,
			Path:   "/auth/me",
			Handler: httputil.NewHandler(func(r httputil.RequestEmpty) (*httputil.Response, error) {
				userID, err := token.Parse(r.Request, jwtSecret)
				if err != nil {
					return nil, &problem.DetailedError{
						Type:   "https://ask-howard.io/problems/unauthorized",
						Title:  "Unauthorized",
						Status: http.StatusUnauthorized,
					}
				}

				user, err := svc.GetByID(r.Context(), userID)
				if err != nil {
					if errors.Is(err, domain.ErrUserNotFound) {
						return nil, &problem.DetailedError{
							Type:   "https://ask-howard.io/problems/unauthorized",
							Title:  "Unauthorized",
							Status: http.StatusUnauthorized,
						}
					}
					return nil, fmt.Errorf("get current user: %w", err)
				}

				return httputil.OK(userResponse{ID: user.ID.String(), Email: user.Email.String()})
			}),
		},
	}
}
