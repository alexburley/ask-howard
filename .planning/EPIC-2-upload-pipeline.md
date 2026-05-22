# Epic 2 — Upload Pipeline

## Goal

Turn an uploaded zip into a set of individually-stored, individually-tracked documents,
without proxying large files through the API. A user requests an upload slot, PUTs the zip
straight to S3, then the backend extracts every entry into its own S3 object and database
row via a durable background job.

## Why

This is the core data-producing flow. The canvas (Epic 3), the detail view (Epic 4), and all
later analysis (Epics 5–7) operate on the per-document rows and objects this epic creates.
Doing extraction in a background job keeps the request fast and survives restarts.

## Outcome

- `document_sets` and `documents` tables, scoped per user.
- River job queue wired in, behind a `JobEnqueuer` outbound port.
- A `DocumentService` with `CreateUploadSlot`, `CompleteUpload`, `ListDocuments`, `GetDocument`.
- A background extraction worker (with zip-bomb guards).
- HTTP endpoints for the upload lifecycle, plus a reusable `currentUserID` auth helper.
- Functional tests for the happy path, ownership scoping, and bad/oversized zips.

## Component stories

- **US-13** — Document schema & migration
- **US-14** — River queue integration (`JobEnqueuer` port)
- **US-15** — sqlc queries, domain types & repository
- **US-16** — `DocumentService` (slot / complete / list / get)
- **US-17** — Zip extraction worker
- **US-18** — Upload HTTP endpoints & `currentUserID` helper

---

## US-13 — Document schema & migration

**As a** developer,
**I want** tables to track upload sets and their extracted documents, scoped per user,
**So that** the system can store metadata and S3 keys for every file.

### Acceptance Criteria

1. `document_sets`: `id`, `user_id` (FK→users, cascade), `original_filename`, `status`
   (`UPLOADING`/`PROCESSING`/`READY`/`FAILED`), `object_key`, `error` (nullable), timestamps.
2. `documents`: `id`, `set_id` (FK→document_sets, cascade), `user_id` (FK→users), `filename`,
   `content_type`, `size_bytes`, `object_key`, `canvas_x` (nullable), `canvas_y` (nullable),
   timestamps.
3. Indexes on `documents(user_id)` and `documents(set_id)`.
4. Schema declared in `schema.hcl`; migration generated via
   `make migrate-diff name=add_documents` and applies cleanly.

### Files

| File | Action |
|------|--------|
| `internal/adapter/outbound/postgres/schema.hcl` | Update |
| `internal/adapter/outbound/postgres/migrations/*_add_documents.sql` | Create (generated) |

---

## US-14 — River queue integration

**As a** backend developer,
**I want** a durable Postgres-backed job queue behind an outbound port,
**So that** extraction (and later analysis) runs reliably in the background without the
service importing River directly.

### Acceptance Criteria

1. River + `riverpgxv5` added; River's migration generated and applied alongside app
   migrations.
2. A River `Client` is constructed in `main.go` and started/stopped with the server.
3. A `JobEnqueuer` outbound port abstracts enqueueing (e.g. `EnqueueExtraction(ctx, setID)`),
   implemented by a thin River-backed adapter.
4. The document service depends on `JobEnqueuer`, never on River types.

### Implementation Notes

- Keep job arg structs in the jobs adapter package; the port speaks domain values (IDs).
- Workers are registered where the River client is built (Epic 2 worker = US-17).

### Files

| File | Action |
|------|--------|
| `internal/port/outbound/jobs.go` | Create — `JobEnqueuer` |
| `internal/adapter/outbound/jobs/` | Create — River client, enqueuer, worker registration |
| `cmd/server/main.go` | Update — construct/start River client |
| `go.mod` / `go.sum` | Update — river |

---

## US-15 — sqlc queries, domain types & repository

**As a** backend developer,
**I want** type-safe queries and a repository for sets and documents,
**So that** services persist and read document data through the outbound port.

### Acceptance Criteria

1. `queries/documents.sql`: create set, set status/error, get set by id+user, insert
   document, list documents by user, get document by id+user, update document position.
2. `make generate` produces `db/` code with no hand edits.
3. `domain.DocumentSet`, `domain.Document`, and sentinel errors
   (`ErrDocumentSetNotFound`, `ErrDocumentNotFound`).
4. `postgres/document.go` implements a `DocumentRepository` outbound interface; maps
   `pgx.ErrNoRows` to the sentinels (mirrors `user.go`).

### Files

| File | Action |
|------|--------|
| `internal/adapter/outbound/postgres/queries/documents.sql` | Create |
| `internal/adapter/outbound/postgres/db/*` | Update (generated) |
| `internal/domain/document.go` | Create |
| `internal/port/outbound/document.go` | Create — `DocumentRepository` |
| `internal/adapter/outbound/postgres/document.go` | Create |

---

## US-16 — `DocumentService`

**As a** user,
**I want** to request an upload slot, signal completion, and list/open my documents,
**So that** I can get files into the system and back out.

### Acceptance Criteria

1. `CreateUploadSlot(ctx, userID, filename)`: inserts a `document_set` (`UPLOADING`),
   presigns a PUT to the set's `object_key`, returns `{ setID, uploadURL, objectKey }`.
2. `CompleteUpload(ctx, userID, setID)`: verifies ownership, marks `PROCESSING`, enqueues an
   extraction job. 404 for another user's set.
3. `ListDocuments(ctx, userID)`: returns the user's documents (with presigned GET URLs).
4. `GetDocument(ctx, userID, docID)`: returns one document + presigned GET URL; 404 if not
   owned.
5. Defined behind `port/inbound.DocumentService`; depends on `DocumentRepository`,
   `ObjectStore`, `JobEnqueuer`.

### Files

| File | Action |
|------|--------|
| `internal/port/inbound/document.go` | Create — `DocumentService` |
| `internal/service/document.go` | Create |

---

## US-17 — Zip extraction worker

**As a** the system,
**I want** to extract an uploaded zip into per-document objects and rows in the background,
**So that** large uploads are processed reliably without blocking the request.

### Acceptance Criteria

1. The worker downloads the set's zip from S3 and streams entries with `archive/zip`.
2. Skips directories and `__MACOSX`/dotfiles; detects `content_type` via
   `gabriel-vasile/mimetype` (already a dependency).
3. Writes each entry to S3 (`sets/{setID}/{uuid}`) and inserts a `documents` row.
4. Enforces a **per-entry size cap and total uncompressed cap** (zip-bomb guard); on breach
   or any error, sets the set to `FAILED` with a message.
5. On success sets the set to `READY`. River retries are bounded.

### Implementation Notes

- Stream rather than buffering whole files where possible; cap readers with `io.LimitReader`.
- Idempotency: a retried job should not duplicate rows (e.g. clear/replace set documents at
  start, or guard on set status).

### Files

| File | Action |
|------|--------|
| `internal/adapter/outbound/jobs/extract.go` | Create |

---

## US-18 — Upload HTTP endpoints & `currentUserID` helper

**As a** frontend client,
**I want** authenticated endpoints to drive the upload lifecycle and poll status,
**So that** the UI can upload a zip and know when documents are ready.

### Acceptance Criteria

1. Endpoints under `/api` (cookie-authenticated):
   - `POST /documents/upload` → create slot, returns upload URL + setID.
   - `POST /documents/sets/{id}/complete` → mark processing + enqueue.
   - `GET /documents/sets/{id}` → set status (for polling).
   - `GET /documents` → list the user's documents.
2. A `handler.currentUserID(r, jwtSecret)` helper extracts the user from the JWT cookie,
   returning a 401 `problem.DetailedError` on failure; `/auth/me` is refactored to use it.
3. Errors use RFC 9457 `problem.DetailedError`; constant strings uppercase.
4. Endpoints registered in `httpserver/server.go`.

### Acceptance Criteria — Tests (US-18 covers Epic 2's functional tests)

5. Functional tests (tagged `functional`, real Postgres + MinIO): upload → complete → poll
   → READY happy path; another user's set returns 404; a malformed/oversized zip ends
   `FAILED`.

### Files

| File | Action |
|------|--------|
| `internal/adapter/inbound/httpserver/handler/document.go` | Create |
| `internal/adapter/inbound/httpserver/handler/auth.go` | Update — extract `currentUserID` |
| `internal/adapter/inbound/httpserver/server.go` | Update — register endpoints |
| `internal/adapter/inbound/httpserver/handler/document_test.go` | Create |

---

## Verification

- `make generate` clean; `make migrate-apply` applies new + River migrations.
- `make test` (functional) green incl. upload lifecycle, ownership 404, zip-bomb → FAILED.
- Manual: via `http/documents.http`, request a slot, PUT a zip to the URL, call complete,
  poll until `READY`, list documents.
- `make lint` clean.
