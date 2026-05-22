package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type DocumentSetStatus string

const (
	DocumentSetStatusUploading  DocumentSetStatus = "UPLOADING"
	DocumentSetStatusProcessing DocumentSetStatus = "PROCESSING"
	DocumentSetStatusReady      DocumentSetStatus = "READY"
	DocumentSetStatusFailed     DocumentSetStatus = "FAILED"
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
	Status           DocumentSetStatus
	ObjectKey        string
	Error            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
