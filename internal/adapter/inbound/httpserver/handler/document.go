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
	DocumentCount    int64  `json:"document_count"`
	Error            string `json:"error,omitempty"`
}

type documentResponse struct {
	ID           string `json:"id"`
	Filename     string `json:"filename"`
	ContentType  string `json:"content_type"`
	SizeBytes    int64  `json:"size_bytes"`
	PresignedURL string `json:"presigned_url"`
}

type setIDParams struct {
	ID uuid.UUID `path:"id" validate:"required"`
}

// DocumentEndpoints returns the document endpoints. All endpoints require a
// valid JWT — apply NewAuthGuard to this group when registering with the server.
func DocumentEndpoints(svc inbound.DocumentService) []httputil.Endpoint {
	return []httputil.Endpoint{
		{
			Method: http.MethodPost,
			Path:   "/documents/upload",
			Handler: httputil.NewHandler(func(r httputil.RequestData[uploadRequestBody]) (*httputil.Response, error) {
				userID, err := auth.UserIDFromContext(r.Context())
				if err != nil {
					return nil, fmt.Errorf("read user from context: %w", err)
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
				userID, err := auth.UserIDFromContext(r.Context())
				if err != nil {
					return nil, fmt.Errorf("read user from context: %w", err)
				}

				set, err := svc.CompleteUpload(r.Context(), r.Params.ID, userID)
				if err != nil {
					if errors.Is(err, domain.ErrDocumentSetNotFound) {
						return nil, notFoundProblem()
					}
					return nil, fmt.Errorf("complete upload: %w", err)
				}

				return httputil.OK(documentSetResponse{
					ID:               set.ID.String(),
					Status:           string(set.Status),
					OriginalFilename: set.OriginalFilename,
					DocumentCount:    0,
				})
			}),
		},
		{
			Method: http.MethodGet,
			Path:   "/documents/sets/{id}",
			Handler: httputil.NewHandler(func(r httputil.RequestParams[setIDParams]) (*httputil.Response, error) {
				userID, err := auth.UserIDFromContext(r.Context())
				if err != nil {
					return nil, fmt.Errorf("read user from context: %w", err)
				}

				result, err := svc.GetDocumentSet(r.Context(), r.Params.ID, userID)
				if err != nil {
					if errors.Is(err, domain.ErrDocumentSetNotFound) {
						return nil, notFoundProblem()
					}
					return nil, fmt.Errorf("get document set: %w", err)
				}

				return httputil.OK(documentSetResponse{
					ID:               result.ID.String(),
					Status:           string(result.Status),
					OriginalFilename: result.OriginalFilename,
					DocumentCount:    result.DocumentCount,
					Error:            result.Error,
				})
			}),
		},
		{
			Method: http.MethodGet,
			Path:   "/documents",
			Handler: httputil.NewHandler(func(r httputil.RequestEmpty) (*httputil.Response, error) {
				userID, err := auth.UserIDFromContext(r.Context())
				if err != nil {
					return nil, fmt.Errorf("read user from context: %w", err)
				}

				docs, err := svc.ListDocuments(r.Context(), userID)
				if err != nil {
					return nil, fmt.Errorf("list documents: %w", err)
				}

				resp := make([]documentResponse, 0, len(docs))
				for i := range docs {
					resp = append(resp, documentResponse{
						ID:           docs[i].ID.String(),
						Filename:     docs[i].Filename,
						ContentType:  docs[i].ContentType,
						SizeBytes:    docs[i].SizeBytes,
						PresignedURL: docs[i].PresignedURL,
					})
				}

				return httputil.OK(resp)
			}),
		},
	}
}

func notFoundProblem() *problem.DetailedError {
	return (&problem.DetailedError{
		Type:   "https://ask-howard.io/problems/not-found",
		Title:  "Not Found",
		Status: http.StatusNotFound,
	}).WithDetail("Document not found")
}
