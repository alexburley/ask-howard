package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alexburley/ask-howard/internal/auth"
	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/inbound"
	"github.com/google/uuid"
	"github.com/nickbryan/httputil"
	"github.com/nickbryan/httputil/problem"
)

type uploadRequestBody struct {
	Filename string `json:"filename" validate:"required"`
}

type uploadSlotResponse struct {
	DocumentSetID string `json:"document_set_id"`
	PresignedURL  string `json:"presigned_url"`
	ObjectKey     string `json:"object_key"`
}

type documentSetResponse struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	OriginalFilename string `json:"original_filename"`
	Error            string `json:"error,omitempty"`
}

type setIDParams struct {
	ID uuid.UUID `path:"id" validate:"required"`
}

func DocumentEndpoints(svc inbound.DocumentService, jwtSecret auth.JWTSecret) []httputil.Endpoint {
	return []httputil.Endpoint{
		{
			Method: http.MethodPost,
			Path:   "/documents/upload",
			Handler: httputil.NewHandler(func(r httputil.RequestData[uploadRequestBody]) (*httputil.Response, error) {
				userID, err := currentUserID(r.Request, jwtSecret)
				if err != nil {
					return nil, err
				}

				result, err := svc.CreateUploadSlot(r.Context(), userID, r.Data.Filename)
				if err != nil {
					return nil, fmt.Errorf("create upload slot: %w", err)
				}

				return httputil.Created(uploadSlotResponse{
					DocumentSetID: result.DocumentSetID.String(),
					PresignedURL:  result.PresignedURL,
					ObjectKey:     result.ObjectKey,
				})
			}),
		},
		{
			Method: http.MethodPost,
			Path:   "/documents/sets/{id}/complete",
			Handler: httputil.NewHandler(func(r httputil.RequestParams[setIDParams]) (*httputil.Response, error) {
				userID, err := currentUserID(r.Request, jwtSecret)
				if err != nil {
					return nil, err
				}

				set, err := svc.CompleteUpload(r.Context(), r.Params.ID, userID)
				if err != nil {
					if errors.Is(err, domain.ErrDocumentSetNotFound) {
						return nil, (&problem.DetailedError{
							Type:   "https://ask-howard.io/problems/not-found",
							Title:  "Not Found",
							Status: http.StatusNotFound,
						}).WithDetail("Document set not found")
					}
					return nil, fmt.Errorf("complete upload: %w", err)
				}

				return httputil.OK(toDocumentSetResponse(&set))
			}),
		},
		{
			Method: http.MethodGet,
			Path:   "/documents/sets/{id}",
			Handler: httputil.NewHandler(func(r httputil.RequestParams[setIDParams]) (*httputil.Response, error) {
				userID, err := currentUserID(r.Request, jwtSecret)
				if err != nil {
					return nil, err
				}

				set, err := svc.GetDocumentSet(r.Context(), r.Params.ID, userID)
				if err != nil {
					if errors.Is(err, domain.ErrDocumentSetNotFound) {
						return nil, (&problem.DetailedError{
							Type:   "https://ask-howard.io/problems/not-found",
							Title:  "Not Found",
							Status: http.StatusNotFound,
						}).WithDetail("Document set not found")
					}
					return nil, fmt.Errorf("get document set: %w", err)
				}

				return httputil.OK(toDocumentSetResponse(&set))
			}),
		},
	}
}

func toDocumentSetResponse(set *domain.DocumentSet) documentSetResponse {
	return documentSetResponse{
		ID:               set.ID.String(),
		Status:           set.Status,
		OriginalFilename: set.OriginalFilename,
		Error:            set.Error,
	}
}
