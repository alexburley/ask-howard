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

var (
	ErrDocumentSetNotFound = errors.New("document set not found")
	ErrDocumentNotFound    = errors.New("document not found")
)

type Document struct {
	ID          uuid.UUID
	SetID       uuid.UUID
	UserID      uuid.UUID
	Filename    string
	ContentType string
	SizeBytes   int64
	ObjectKey   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

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
