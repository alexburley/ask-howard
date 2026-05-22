# Epic 2 — Document Canvas

## Goal

A user can see all their extracted documents floating on an interactive canvas, and arrange
them spatially — with the layout persisting across sessions.

## Why

The canvas is the product's primary surface. It transforms a list of files into a spatial
workspace where patterns, families, and timelines can emerge visually. Built on
`@xyflow/react`, it grows into the relationship-graph epic (edges between nodes) without a
rewrite.

## Outcome

- Documents appear as nodes on a pan/zoom canvas immediately after extraction is complete.
- Images render as thumbnails; non-images show a type icon and filename.
- Nodes are draggable; positions persist to the backend and survive page reload.
- Full functional test coverage of the position endpoint; e2e tests covering the canvas flow.

---

## US-13 — Browse documents on the canvas

**As a** user,
**I want** to see all my documents laid out on a scrollable, zoomable canvas,
**So that** I can get a spatial overview of my collection at a glance.

### Acceptance Criteria

**Backend:**

1. `GET /api/documents` (already delivered in US-12) returns per-document `canvas_x`,
   `canvas_y`, `presign_url` (15-minute GET URL), `filename`, `content_type`, `size_bytes`.
2. Documents with `null` positions receive a deterministic scatter layout server-side or
   client-side grid so they are never stacked on top of each other.

**Frontend:**

3. `DocumentCanvas` uses `@xyflow/react` and renders one node per document.
4. Image documents (`image/*`) render a downscaled `<img>` loaded from `presign_url`.
5. Non-image documents show a file-type icon and the filename.
6. Pan and zoom work. The initial viewport fits all nodes.
7. `Workspace` renders the canvas below the upload control once at least one document
   exists for the user.

**E2E test** (`web/e2e/canvas.spec.ts`):

8. After uploading a fixture zip (reusing the upload helper from US-12's test), the canvas
   renders with the expected number of nodes.
9. At least one image node has a visible `<img>` element.

### Out of Scope

- Selecting a node / opening detail (US-15).
- Persisting drag position (US-14).
- Relationship edges (deferred Epic 5).

### Implementation Notes

- Add `@xyflow/react` to `web/package.json`.
- Presigned URLs expire; the frontend should re-fetch documents (and hence fresh URLs) on
  focus/visibility change or after 10 minutes.
- Default layout for new documents: simple grid with fixed column width + row height using
  the document index. Stored `canvas_x`/`canvas_y` override this on subsequent loads.

### Files

| File | Action |
|------|--------|
| `web/package.json` | Update — `@xyflow/react` |
| `web/src/components/DocumentCanvas.tsx` | Create |
| `web/src/components/DocumentNode.tsx` | Create |
| `web/src/hooks/useDocuments.ts` | Update — include presign URLs, expose as xyflow nodes |
| `web/src/pages/Workspace.tsx` | Update — render canvas |
| `web/src/index.css` | Update — canvas container styles |
| `web/e2e/canvas.spec.ts` | Create |

---

## US-14 — Arrange documents on the canvas

**As a** user,
**I want** to drag my documents into a layout that makes sense to me, and have that
arrangement be there when I come back,
**So that** I can organise my collection spatially without losing my work.

### Acceptance Criteria

**Backend:**

1. `PATCH /api/documents/{id}/position` (authenticated): updates `canvas_x`, `canvas_y`
   for the given document. Verifies ownership; returns 404 for another user's document.
2. Returns 200 with the updated document or an RFC 9457 error.

**Functional tests** (`//go:build functional`, `make test`):

3. `PATCH /documents/{id}/position` persists and is returned by `GET /documents`.
4. Patching another user's document returns 404.
5. Unauthenticated request returns 401.

**Frontend:**

6. Nodes are draggable via `@xyflow/react`'s built-in drag support.
7. On drag-stop, the new `(x, y)` is sent to `PATCH /documents/{id}/position`, debounced
   500 ms to avoid a request per animation frame.
8. On reload, documents render at their saved positions (not the default grid).

**E2E test** (`web/e2e/canvas.spec.ts`, extends US-13):

9. Drags a node to a new position, reloads the page, confirms the node is at the new
   position (not the default grid position).

### Out of Scope

- Multi-select / group drag.
- Undo / redo.

### Implementation Notes

- `PATCH` handler: single sqlc `UpdateDocumentPosition` query.
- Debounce in the frontend hook (not component) so it's testable.
- xyflow stores positions in its own node state; sync back to API on `onNodeDragStop`.

### Files

| File | Action |
|------|--------|
| `internal/adapter/outbound/postgres/queries/documents.sql` | Update — `UpdateDocumentPosition` |
| `internal/adapter/outbound/postgres/db/*` | Update (generated) |
| `internal/service/document.go` | Update — `UpdatePosition` |
| `internal/adapter/outbound/postgres/document.go` | Update — position repo method |
| `internal/adapter/inbound/httpserver/handler/document.go` | Update — PATCH handler |
| `internal/adapter/inbound/httpserver/handler/document_test.go` | Update (functional) |
| `internal/adapter/inbound/httpserver/server.go` | Update — register PATCH endpoint |
| `web/src/documents/api.ts` | Update — `updatePosition` |
| `web/src/hooks/useDocuments.ts` | Update — debounced position persist |
| `web/src/components/DocumentCanvas.tsx` | Update — `onNodeDragStop` callback |
| `web/e2e/canvas.spec.ts` | Update — drag + reload assertion |

---

## Verification

- `make test` (functional) green: position update, ownership 404, 401.
- `make lint` clean.
- `make e2e` green: canvas renders nodes; drag + reload preserves position.
- `npx tsc --noEmit` clean.
