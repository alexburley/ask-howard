package domain

import "time"

type ClaimStatus string

const (
	ClaimStatusSuggested         ClaimStatus = "suggested"
	ClaimStatusConfirmed         ClaimStatus = "confirmed"
	ClaimStatusRejected          ClaimStatus = "rejected"
	ClaimStatusDisputed          ClaimStatus = "disputed"
	ClaimStatusNeedsMoreEvidence ClaimStatus = "needs_more_evidence"
)

type Claim struct {
	ID           string
	SubjectType  string
	SubjectID    string
	ClaimType    string
	Value        string
	SourceDocID  string
	EvidenceText string
	Confidence   float64
	Status       ClaimStatus
	CreatedBy    string
	ReviewedAt   *time.Time
}
