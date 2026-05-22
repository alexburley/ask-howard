# Epic 3 — Canvas Workspace

## Goal

Replace the placeholder `Workspace` with an interactive canvas where a user's documents
float as draggable nodes. Users upload a zip from here, watch it process, and rearrange
documents on an infinite pan/zoom surface whose layout persists.

## Why

The canvas is the product's primary surface. Built on `@xyflow/react` (nodes now, edges
later), it grows directly into the relationship-graph epic without a rewrite.

## Outcome

- `@xyflow/react` added; a documents API client + `useDocuments` hook mirroring the existing
  `auth` patterns.
- An upload control with progress and processing states.
- A `DocumentCanvas` rendering one node per document, with pan/zoom and drag-to-reposition
  persisted to the backend.

## Component stories

- **US-19** — Documents API client
- **US-20** — `useDocuments` hook
- **US-21** — Upload control (drop zone + progress)
- **US-22** — Document canvas (xyflow)
- **US-23** — Persist canvas layout

---

## US-19 — Documents API client

**As a** frontend developer,
**I want** a typed client for the document endpoints (and direct-to-S3 upload),
**So that** components never touch raw `fetch` or leak network errors.

### Acceptance Criteria

1. `web/src/documents/api.ts`: `requestUploadSlot`, `uploadToPresignedUrl` (PUT to S3 with
   progress callback), `completeUpload`, `pollSet`, `listDocuments`, `getDocument`.
2. Throws typed errors mirroring `auth/api.ts`'s `AuthError` style (typed `code` field); no
   raw fetch errors surfaced.
3. `uploadToPresignedUrl` uses `XMLHttpRequest` (or fetch + stream) to report upload progress.
4. `npx tsc --noEmit` clean.

### Files

| File | Action |
|------|--------|
| `web/src/documents/api.ts` | Create |
| `web/src/documents/types.ts` | Create |

---

## US-20 — `useDocuments` hook

**As a** frontend developer,
**I want** a hook that owns document state and the upload lifecycle,
**So that** the canvas and upload control share one source of truth.

### Acceptance Criteria

1. `web/src/hooks/useDocuments.ts` fetches the document list on mount and exposes it.
2. Exposes `upload(file)` that runs request-slot → PUT (with progress) → complete → poll
   set status, then refreshes the list when `READY`.
3. Exposes progress and a per-upload status (`uploading`/`processing`/`ready`/`failed`).
4. Exposes `updatePosition(docId, x, y)` (used by US-23).

### Files

| File | Action |
|------|--------|
| `web/src/hooks/useDocuments.ts` | Create |

---

## US-21 — Upload control

**As a** user,
**I want** to drop or pick a zip and see its upload + processing progress,
**So that** I know my documents are being ingested.

### Acceptance Criteria

1. A drop zone / "Upload zip" button in the workspace (accepts `.zip`).
2. Shows upload progress (%) then a "Processing…" state while extraction runs.
3. On `READY` the new documents appear on the canvas; on `FAILED` a clear error is shown.
4. Rejects non-zip files client-side with a friendly message.

### Files

| File | Action |
|------|--------|
| `web/src/components/UploadControl.tsx` | Create |
| `web/src/index.css` | Update — upload/progress styles |

---

## US-22 — Document canvas

**As a** user,
**I want** my documents to float on a pan/zoom canvas,
**So that** I can browse and spatially arrange them.

### Acceptance Criteria

1. `DocumentCanvas` uses `@xyflow/react`; one node per document.
2. Image documents render a downscaled thumbnail from the presigned GET URL; non-images show
   a type icon + filename.
3. Pan and zoom work; nodes are draggable.
4. Documents with no stored position get a sensible initial scatter/grid layout.
5. Selecting a node surfaces selection state (consumed by Epic 4's detail view).

### Files

| File | Action |
|------|--------|
| `web/src/components/DocumentCanvas.tsx` | Create |
| `web/src/components/DocumentNode.tsx` | Create |
| `web/src/pages/Workspace.tsx` | Update — render upload control + canvas |
| `web/package.json` | Update — `@xyflow/react` |

---

## US-23 — Persist canvas layout

**As a** user,
**I want** my document arrangement to be remembered,
**So that** the canvas looks the same when I return.

### Acceptance Criteria

1. On drag-stop, the node's `(x, y)` is sent to the backend (debounced).
2. Backend exposes `PATCH /documents/{id}/position` (`canvas_x`, `canvas_y`), owner-scoped,
   backed by an sqlc update query.
3. Positions survive reload — documents render where they were left.

### Files

| File | Action |
|------|--------|
| `internal/adapter/inbound/httpserver/handler/document.go` | Update — position endpoint |
| `internal/adapter/outbound/postgres/queries/documents.sql` | Update — update-position query |
| `web/src/documents/api.ts` | Update — `updatePosition` |
| `web/src/hooks/useDocuments.ts` | Update — debounced persist |

---

## Verification

- `npx tsc --noEmit` clean.
- In the browser: login → upload a real image zip → watch progress/processing → documents
  appear → drag one → reload → it stays put.
