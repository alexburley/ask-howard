package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexburley/ask-howard/internal/adapter/outbound/postgres/db"
	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/outbound"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentRepository struct {
	queries *db.Queries
}

var _ outbound.DocumentRepository = (*DocumentRepository)(nil)

func NewDocumentRepository(pool *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{queries: db.New(pool)}
}

func (r *DocumentRepository) CreateDocumentSet(ctx context.Context, params outbound.CreateDocumentSetParams) (domain.DocumentSet, error) {
	row, err := r.queries.CreateDocumentSet(ctx, db.CreateDocumentSetParams{
		UserID:           params.UserID,
		OriginalFilename: params.OriginalFilename,
		Status:           params.Status,
		ObjectKey:        params.ObjectKey,
	})
	if err != nil {
		return domain.DocumentSet{}, fmt.Errorf("create document set: %w", err)
	}
	return toDomainDocumentSet(&row), nil
}

func (r *DocumentRepository) GetDocumentSetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (domain.DocumentSet, error) {
	row, err := r.queries.GetDocumentSetByIDAndUser(ctx, db.GetDocumentSetByIDAndUserParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.DocumentSet{}, domain.ErrDocumentSetNotFound
		}
		return domain.DocumentSet{}, fmt.Errorf("get document set: %w", err)
	}
	return toDomainDocumentSet(&row), nil
}

func (r *DocumentRepository) UpdateDocumentSetStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) (domain.DocumentSet, error) {
	row, err := r.queries.UpdateDocumentSetStatus(ctx, db.UpdateDocumentSetStatusParams{
		ID:     id,
		Status: status,
		Error:  pgtype.Text{String: errorMsg, Valid: errorMsg != ""},
	})
	if err != nil {
		return domain.DocumentSet{}, fmt.Errorf("update document set status: %w", err)
	}
	return toDomainDocumentSet(&row), nil
}

func (r *DocumentRepository) InsertDocument(ctx context.Context, params *outbound.InsertDocumentParams) (domain.Document, error) {
	row, err := r.queries.InsertDocument(ctx, db.InsertDocumentParams{
		SetID:       params.SetID,
		UserID:      params.UserID,
		Filename:    params.Filename,
		ContentType: params.ContentType,
		SizeBytes:   params.SizeBytes,
		ObjectKey:   params.ObjectKey,
	})
	if err != nil {
		return domain.Document{}, fmt.Errorf("insert document: %w", err)
	}
	return toDomainDocument(&row), nil
}

func (r *DocumentRepository) ListDocumentsByUser(ctx context.Context, userID uuid.UUID) ([]domain.Document, error) {
	rows, err := r.queries.ListDocumentsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list documents by user: %w", err)
	}
	docs := make([]domain.Document, len(rows))
	for i := range rows {
		docs[i] = toDomainDocument(&rows[i])
	}
	return docs, nil
}

func (r *DocumentRepository) GetDocumentByIDAndUser(ctx context.Context, id, userID uuid.UUID) (domain.Document, error) {
	row, err := r.queries.GetDocumentByIDAndUser(ctx, db.GetDocumentByIDAndUserParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Document{}, domain.ErrDocumentNotFound
		}
		return domain.Document{}, fmt.Errorf("get document: %w", err)
	}
	return toDomainDocument(&row), nil
}

func (r *DocumentRepository) DeleteDocumentsBySetID(ctx context.Context, setID uuid.UUID) error {
	if err := r.queries.DeleteDocumentsBySetID(ctx, setID); err != nil {
		return fmt.Errorf("delete documents by set: %w", err)
	}
	return nil
}

func (r *DocumentRepository) CountDocumentsBySetID(ctx context.Context, setID uuid.UUID) (int64, error) {
	count, err := r.queries.CountDocumentsBySetID(ctx, setID)
	if err != nil {
		return 0, fmt.Errorf("count documents by set: %w", err)
	}
	return count, nil
}

func toDomainDocument(d *db.Document) domain.Document {
	return domain.Document{
		ID:          d.ID,
		SetID:       d.SetID,
		UserID:      d.UserID,
		Filename:    d.Filename,
		ContentType: d.ContentType,
		SizeBytes:   d.SizeBytes,
		ObjectKey:   d.ObjectKey,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func toDomainDocumentSet(ds *db.DocumentSet) domain.DocumentSet {
	return domain.DocumentSet{
		ID:               ds.ID,
		UserID:           ds.UserID,
		OriginalFilename: ds.OriginalFilename,
		Status:           ds.Status,
		ObjectKey:        ds.ObjectKey,
		Error:            ds.Error.String,
		CreatedAt:        ds.CreatedAt,
		UpdatedAt:        ds.UpdatedAt,
	}
}
