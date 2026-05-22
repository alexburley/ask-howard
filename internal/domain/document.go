package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	DocumentSetStatusUploading  = "UPLOADING"
	DocumentSetStatusProcessing = "PROCESSING"
	DocumentSetStatusReady      = "READY"
	DocumentSetStatusFailed     = "FAILED"
)

var ErrDocumentSetNotFound = errors.New("document set not found")

type DocumentSet struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	OriginalFilename string
	Status           string
	ObjectKey        string
	Error            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
