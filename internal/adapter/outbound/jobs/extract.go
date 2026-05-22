package jobs

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/alexburley/ask-howard/internal/domain"
	"github.com/alexburley/ask-howard/internal/port/outbound"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

const (
	DefaultMaxEntryBytes int64 = 100 * 1024 * 1024 // 100 MB per entry
	DefaultMaxTotalBytes int64 = 500 * 1024 * 1024 // 500 MB total uncompressed
)

var (
	errEntryTooLarge = errors.New("entry exceeds per-file size limit")
	errTotalTooLarge = errors.New("total uncompressed size exceeds limit")
)

type ExtractionArgs struct {
	SetID  uuid.UUID `json:"set_id"`
	UserID uuid.UUID `json:"user_id"`
}

func (ExtractionArgs) Kind() string { return "extraction" }

type ExtractionWorker struct {
	river.WorkerDefaults[ExtractionArgs]
	store         outbound.ObjectStore
	docs          outbound.DocumentRepository
	maxEntryBytes int64
	maxTotalBytes int64
}

var _ river.Worker[ExtractionArgs] = (*ExtractionWorker)(nil)

func NewExtractionWorker(store outbound.ObjectStore, docs outbound.DocumentRepository) *ExtractionWorker {
	return &ExtractionWorker{
		store:         store,
		docs:          docs,
		maxEntryBytes: DefaultMaxEntryBytes,
		maxTotalBytes: DefaultMaxTotalBytes,
	}
}

func NewExtractionWorkerWithLimits(store outbound.ObjectStore, docs outbound.DocumentRepository, maxEntry, maxTotal int64) *ExtractionWorker {
	return &ExtractionWorker{
		store:         store,
		docs:          docs,
		maxEntryBytes: maxEntry,
		maxTotalBytes: maxTotal,
	}
}

func (w *ExtractionWorker) Work(ctx context.Context, job *river.Job[ExtractionArgs]) error {
	setID := job.Args.SetID
	userID := job.Args.UserID

	set, err := w.docs.GetDocumentSetByIDAndUser(ctx, setID, userID)
	if err != nil {
		return fmt.Errorf("get document set: %w", err)
	}

	if err := w.docs.DeleteDocumentsBySetID(ctx, setID); err != nil {
		return fmt.Errorf("delete partial documents: %w", err)
	}

	if extractErr := w.extract(ctx, &set); extractErr != nil {
		if _, updateErr := w.docs.UpdateDocumentSetStatus(ctx, setID, domain.DocumentSetStatusFailed, extractErr.Error()); updateErr != nil {
			return fmt.Errorf("set failed status: %w", updateErr)
		}
		return nil
	}

	if _, err := w.docs.UpdateDocumentSetStatus(ctx, setID, domain.DocumentSetStatusReady, ""); err != nil {
		return fmt.Errorf("set ready status: %w", err)
	}

	return nil
}

func (w *ExtractionWorker) extract(ctx context.Context, set *domain.DocumentSet) error {
	rc, err := w.store.Get(ctx, set.ObjectKey)
	if err != nil {
		return fmt.Errorf("get zip from storage: %w", err)
	}
	defer func() { _ = rc.Close() }()

	raw, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("read zip: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	var totalBytes int64
	for _, f := range zr.File {
		if shouldSkip(f.Name) {
			continue
		}

		if int64(f.UncompressedSize64) > w.maxEntryBytes {
			return fmt.Errorf("%w: %q", errEntryTooLarge, f.Name)
		}

		totalBytes += int64(f.UncompressedSize64)
		if totalBytes > w.maxTotalBytes {
			return errTotalTooLarge
		}

		if err := w.extractEntry(ctx, set, f); err != nil {
			return err
		}
	}

	return nil
}

func (w *ExtractionWorker) extractEntry(ctx context.Context, set *domain.DocumentSet, f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open zip entry %q: %w", f.Name, err)
	}
	defer func() { _ = rc.Close() }()

	limited := io.LimitReader(rc, w.maxEntryBytes)
	data, err := io.ReadAll(limited)
	if err != nil {
		return fmt.Errorf("read zip entry %q: %w", f.Name, err)
	}

	ct := mimetype.Detect(data).String()
	key := fmt.Sprintf("sets/%s/%s", set.ID, uuid.New())

	if err := w.store.Put(ctx, key, bytes.NewReader(data), int64(len(data)), ct); err != nil {
		return fmt.Errorf("put document %q: %w", f.Name, err)
	}

	_, err = w.docs.InsertDocument(ctx, &outbound.InsertDocumentParams{
		SetID:       set.ID,
		UserID:      set.UserID,
		Filename:    f.Name,
		ContentType: ct,
		SizeBytes:   int64(len(data)),
		ObjectKey:   key,
	})
	if err != nil {
		return fmt.Errorf("insert document %q: %w", f.Name, err)
	}

	return nil
}

func shouldSkip(name string) bool {
	if strings.HasSuffix(name, "/") {
		return true
	}
	base := name
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		base = name[idx+1:]
	}
	if strings.HasPrefix(base, ".") {
		return true
	}
	if strings.Contains(name, "__MACOSX") {
		return true
	}
	return false
}
