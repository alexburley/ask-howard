# Epic 1 — Document Upload

## Goal

A user can upload a zip of ancestry documents and watch it be processed into individually
accessible files, without ever waiting on a slow server-side proxy.

## Why

This is the core value-creation step. Nothing else in the product — canvas, analysis,
knowledge base — exists without documents in the system. Getting this right (large-file
safe, durable background extraction, per-user scoping) unblocks every subsequent epic.

## Outcome

- A user can pick or drop a zip, see it upload directly to storage, watch per-set status
  move through `UPLOADING → PROCESSING → READY`, and know when their documents are available.
- A `documents` and `document_sets` data model scoped per user.
- An extraction worker that safely handles large/adversarial zips.
- Full functional test coverage of the upload lifecycle and extraction edge cases.
- An e2e test covering the end-to-end upload flow in the browser.

---

## ~~US-10 — Local document storage~~ ✓ Done

Archived → [`.archive/EPIC-1/US-10-local-document-storage.md`](../.archive/EPIC-1/US-10-local-document-storage.md)
Commit: `1acce9a`

---

## ~~US-11 — Upload a document set~~ ✓ Done

Archived → [`.archive/EPIC-1/US-11-upload-a-document-set.md`](../.archive/EPIC-1/US-11-upload-a-document-set.md)

---

## US-12 — Extract and track document processing

**As a** user,
**I want** to see my zip be extracted into individual documents and know when they are ready,
**So that** I can follow progress and trust that all my files made it in.

### Acceptance Criteria

**Backend:**

1. A River job (`ExtractionJob`) is enqueued by `CompleteUpload`. It downloads the zip
   from S3, streams entries via `archive/zip`, writes each to S3 as `sets/{setID}/{uuid}`,
   and inserts a `documents` row per entry.
2. On success the set transitions to `READY`; on error to `FAILED` with a message stored
   in `documents_sets.error`.
3. Skips directories, `__MACOSX`, and dotfiles. Detects `content_type` via
   `gabriel-vasile/mimetype`.
4. Enforces a **per-entry size cap** and **total uncompressed cap** (zip-bomb guard). A
   breach transitions the set to `FAILED`.
5. `GET /api/documents/sets/{id}` (authenticated) returns `{ status, documentCount, error }`.
   Returns 404 for another user's set.
6. `GET /api/documents` (authenticated) returns the user's documents with presigned GET
   URLs (15-minute expiry).

**Functional tests** (`//go:build functional`, `make test`):

7. A valid zip → extraction completes → set becomes `READY` → documents are listed.
8. A zip with a zip-bomb entry (uncompressed > cap) → set becomes `FAILED`.
9. A corrupt / non-zip file → set becomes `FAILED` with an error message.
10. `GET /sets/{id}` for another user's set returns 404.
11. `GET /documents` only returns the authenticated user's documents.

**Frontend:**

12. After `/complete`, the upload hook polls `GET /sets/{id}` every 2 s until `READY` or
    `FAILED`.
13. While polling: "Processing your documents…" indicator in `Workspace`.
14. On `READY`: success state displayed; `useDocuments` hook refreshes the document list.
15. On `FAILED`: error message displayed with a retry option (re-upload).

**E2E test** (`web/e2e/upload.spec.ts`, extends US-11 test):

16. Uploads a fixture zip → watches "Processing" → confirms documents appear (count > 0)
    before navigating to the canvas.

### Out of Scope

- Rendering documents on the canvas (US-13).
- Presigned URLs for preview (consumed by US-13/US-15).

### Implementation Notes

- River: `riverqueue/river` + `riverpgxv5`. River's own migration generated alongside app
  migrations. Worker registered where the River client is built; the `DocumentService`
  depends on a `JobEnqueuer` outbound port so it never imports River types directly.
- `documents` table: `id`, `set_id` FK, `user_id` FK, `filename`, `content_type`,
  `size_bytes`, `object_key`, `canvas_x`/`canvas_y` (nullable), timestamps.
- Use `io.LimitReader` per entry for the size cap.
- Idempotent retry: at job start, delete any partial documents for the set before
  re-extracting (safe because the set is in `PROCESSING`, not `READY`).

### Files

| File | Action |
|------|--------|
| `internal/adapter/outbound/postgres/schema.hcl` | Update — `documents` table |
| `internal/adapter/outbound/postgres/migrations/*` | Create (generated) |
| `internal/adapter/outbound/postgres/queries/documents.sql` | Update — insert doc, list by user, get by id+user |
| `internal/adapter/outbound/postgres/db/*` | Update (generated) |
| `internal/domain/document.go` | Update — `Document` type, `ErrDocumentNotFound` |
| `internal/port/outbound/document.go` | Update — `DocumentRepository` (docs) |
| `internal/port/outbound/jobs.go` | Create — `JobEnqueuer` |
| `internal/service/document.go` | Update — `ListDocuments`, `GetDocument` |
| `internal/adapter/outbound/jobs/extract.go` | Create — River worker |
| `internal/adapter/outbound/jobs/client.go` | Create — River client + enqueuer |
| `internal/adapter/outbound/postgres/document.go` | Update — doc repository methods |
| `internal/adapter/inbound/httpserver/handler/document.go` | Update — status poll + list |
| `internal/adapter/inbound/httpserver/handler/document_test.go` | Update (functional) |
| `internal/adapter/inbound/httpserver/server.go` | Update — register list endpoint |
| `cmd/server/main.go` | Update — construct/start River client |
| `go.mod` / `go.sum` | Update — river |
| `web/src/documents/api.ts` | Update — `pollSet`, `listDocuments` |
| `web/src/hooks/useUpload.ts` | Update — polling logic |
| `web/src/hooks/useDocuments.ts` | Create — document list state |
| `web/src/pages/Workspace.tsx` | Update — processing/ready/error states |
| `web/e2e/upload.spec.ts` | Update — assert document count after READY |

---

## Verification

- `make test` (functional) green: S3 adapter round-trip, upload lifecycle, extraction happy
  path + failure modes, ownership scoping.
- `make lint` clean.
- `make e2e` green: upload → processing → ready.
- Manual: `http/documents.http` — slot → PUT → complete → poll → list documents.
