package domain

import "time"

type (
	OCRStatus      string
	AnalysisStatus string
)

const (
	OCRStatusPending    OCRStatus = "pending"
	OCRStatusProcessing OCRStatus = "processing"
	OCRStatusCompleted  OCRStatus = "completed"
	OCRStatusFailed     OCRStatus = "failed"
)

const (
	AnalysisStatusPending    AnalysisStatus = "pending"
	AnalysisStatusProcessing AnalysisStatus = "processing"
	AnalysisStatusCompleted  AnalysisStatus = "completed"
	AnalysisStatusFailed     AnalysisStatus = "failed"
)

type Document struct {
	ID             string
	Title          string
	DocumentType   string
	FileURL        string
	UploadedAt     time.Time
	ApproxDate     *time.Time
	SourceNotes    string
	OCRStatus      OCRStatus
	AnalysisStatus AnalysisStatus
}
