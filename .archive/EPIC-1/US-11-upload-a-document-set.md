# US-11 — Upload a document set ✓ Done

**Archived from:** `.planning/EPIC-1-document-upload.md`
**Commit:** `HEAD` (see git log for hash)

## What was delivered

Full-stack upload slot flow: browser drop zone → presigned PUT to MinIO → complete notification → PROCESSING status.

### Backend
- `internal/domain/document.go` — `DocumentSet` type, `UPLOADING/PROCESSING/READY/FAILED` constants, `ErrDocumentSetNotFound`
- `internal/port/outbound/document.go` — `DocumentRepository` interface (sets only)
- `internal/port/inbound/document.go` — `DocumentService` interface + `UploadSlotResult`
- `internal/adapter/outbound/postgres/document.go` — sqlc-backed Postgres repo
- `internal/service/document.go` — `CreateUploadSlot`, `CompleteUpload`, `GetDocumentSet`
- `internal/adapter/inbound/httpserver/handler/helpers.go` — `currentUserID` helper extracted from auth handler
- `internal/adapter/inbound/httpserver/handler/document.go` — `POST /api/documents/upload`, `POST /api/documents/sets/{id}/complete`, `GET /api/documents/sets/{id}`
- `internal/adapter/inbound/httpserver/server.go` — registers document endpoints when `docSvc != nil`
- `cmd/server/main.go` — wires `DocumentRepository` + `DocumentService`
- `http/documents.http` — manual request file

### Functional tests (`//go:build functional`)
- Upload slot returns presigned URL
- Unauthenticated request returns 401
- Complete transitions set to PROCESSING
- Complete with another user's set ID returns 404
- Get returns UPLOADING status + correct filename
- Get with another user's set ID returns 404

### Frontend
- `web/src/documents/types.ts` + `web/src/documents/api.ts` — typed client using XHR for upload progress
- `web/src/hooks/useUpload.ts` — idle → uploading (with progress %) → processing → done/error
- `web/src/components/UploadControl.tsx` — drop zone + click-to-browse + progress bar + error/done states
- `web/src/pages/Workspace.tsx` — renders `UploadControl`
- `web/src/index.css` — upload/progress bar styles
- `web/e2e/upload.spec.ts` — registers, uploads fixture zip, asserts upload/processing state visible
- `web/e2e/fixtures/test.zip` — minimal valid empty zip fixture

## Implementation notes

- `S3_PRESIGN_ENDPOINT` env var allows presigned URLs to use `localhost:9000` (browser-accessible) while the API talks to `minio:9000` internally.
- `DocumentEndpoints` only registered when `docSvc != nil` — existing auth/health test suites pass `nil` and are unaffected.
- `wrapcheck` linter requires errors from `DocumentRepository.GetDocumentSetByIDAndUser` to be wrapped in `GetDocumentSet`; done with `fmt.Errorf("get document set: %w", err)`.
