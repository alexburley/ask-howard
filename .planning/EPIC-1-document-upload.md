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

## US-11 — Upload a document set

**As a** user,
**I want** to select a zip file, have it upload directly to storage, and see clear progress,
**So that** I can get my documents into the system without waiting for a slow server upload.

### Acceptance Criteria

**Backend:**

1. `POST /api/documents/upload` (authenticated): inserts a `document_sets` row
   (`status = UPLOADING`, `user_id`, `original_filename`), presigns a PUT URL for the
   zip object, and returns `{ setId, uploadUrl }`.
2. `POST /api/documents/sets/{id}/complete` (authenticated): verifies ownership, marks
   `PROCESSING`, returns `{ status: "PROCESSING" }`. Returns 404 for another user's set.
3. Both endpoints use the `currentUserID` helper (extracted from JWT cookie; returns 401 on
   failure). `handler/auth.go` is refactored to use the same helper.
4. Errors use RFC 9457 `problem.DetailedError`; string values uppercase.

**Functional tests** (`//go:build functional`, `make test`):

5. `POST /upload` returns a presigned URL; the URL accepts an HTTP PUT.
6. `POST /sets/{id}/complete` transitions status to `PROCESSING`.
7. Calling `/complete` with another user's set ID returns 404.
8. Unauthenticated requests to both endpoints return 401.

**Frontend:**

9. An "Upload documents" drop zone / file picker in `Workspace` accepts `.zip` only.
10. Requests a slot, PUTs the zip to the presigned URL with `XMLHttpRequest` for progress,
    then calls `/complete`.
11. Shows a progress bar (0–100%) during the PUT phase.
12. After `/complete` transitions to a "Processing…" state (handled in US-12).
13. Rejects non-zip files with a friendly client-side message before any network call.

**E2E test** (`web/e2e/upload.spec.ts`, `make e2e`):

14. Registers a fresh user, uploads a small fixture zip, confirms progress reaches 100%
    and the processing state appears.

### Out of Scope

- Extracting documents (US-12).
- Showing documents on the canvas (US-13).
- CORS headers beyond what MinIO requires for the presigned PUT.

### Implementation Notes

- `document_sets` table: `id`, `user_id` FK→users, `original_filename`, `status`
  (`UPLOADING`/`PROCESSING`/`READY`/`FAILED`), `object_key`, `error` (nullable), timestamps.
- Schema in `schema.hcl`; migration via `make migrate-diff name=add_document_sets`.
- `currentUserID(r *http.Request, secret auth.JWTSecret) (uuid.UUID, error)` — shared
  helper in the handler package, used by the new document handlers and `/auth/me`.
- The `documents/api.ts` client uses `XMLHttpRequest` for upload so `onprogress` fires.

### Files

| File | Action |
|------|--------|
| `internal/adapter/outbound/postgres/schema.hcl` | Update — `document_sets` table |
| `internal/adapter/outbound/postgres/migrations/*` | Create (generated) |
| `internal/adapter/outbound/postgres/queries/documents.sql` | Create — set queries |
| `internal/adapter/outbound/postgres/db/*` | Update (generated via `make generate`) |
| `internal/domain/document.go` | Create — `DocumentSet`, sentinel errors |
| `internal/port/outbound/document.go` | Create — `DocumentRepository` (sets) |
| `internal/port/inbound/document.go` | Create — `DocumentService` |
| `internal/adapter/outbound/postgres/document.go` | Create — set repository |
| `internal/service/document.go` | Create — `CreateUploadSlot`, `CompleteUpload` |
| `internal/adapter/inbound/httpserver/handler/document.go` | Create — upload + complete handlers |
| `internal/adapter/inbound/httpserver/handler/document_test.go` | Create (functional) |
| `internal/adapter/inbound/httpserver/handler/auth.go` | Update — extract `currentUserID` |
| `internal/adapter/inbound/httpserver/server.go` | Update — register endpoints |
| `web/src/documents/api.ts` | Create — `requestUploadSlot`, `uploadToPresignedUrl`, `completeUpload` |
| `web/src/documents/types.ts` | Create |
| `web/src/hooks/useUpload.ts` | Create — upload lifecycle with progress |
| `web/src/components/UploadControl.tsx` | Create — drop zone, progress bar |
| `web/src/pages/Workspace.tsx` | Update — render upload control |
| `web/src/index.css` | Update — upload/progress styles |
| `web/e2e/upload.spec.ts` | Create |
| `http/documents.http` | Create — manual request file |

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
