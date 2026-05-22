# Epic 3 — Document Detail

## Goal

A user can click any document on the canvas to open a detail view: see its metadata, preview
its content inline, and download the original file.

## Why

Closes the MVP loop: upload → browse → **read**. It also fixes the integration seam for
on-demand AI tagging (Epic 4), so adding analysis later is additive rather than a redesign.

## Outcome

- Clicking a canvas node opens a detail panel without leaving the canvas.
- Images render inline; PDFs embed; other file types offer a download link.
- All previews use short-lived presigned GET URLs fetched fresh on open.
- A visible-but-empty Tags section marks the Epic 4 seam.
- e2e test covering open, preview, and download.

---

## US-15 — Open, preview, and download a document

**As a** user,
**I want** to click a document and immediately see a preview, its metadata, and a download
link,
**So that** I can read and save my ancestry documents without leaving the app.

### Acceptance Criteria

**Backend:**

1. `GET /api/documents/{id}` (authenticated): returns document metadata + a fresh
   presigned GET URL (15-minute expiry). Returns 404 for another user's document.

**Functional tests** (`//go:build functional`, `make test`):

2. `GET /documents/{id}` returns correct metadata and a non-empty presigned URL.
3. Another user's document returns 404.
4. Unauthenticated request returns 401.

**Frontend:**

5. Clicking a canvas node opens `DocumentDetail` as a side panel (canvas remains visible
   and interactive underneath).
6. The panel shows: filename, content type, human-readable file size, upload date.
7. Images (`image/*`) render inline via `<img src={presignUrl}>`.
8. PDFs (`application/pdf`) render via `<iframe src={presignUrl}>`.
9. All other types show a file-type icon, filename, and a "Download" button.
10. A "Download" button/link is present on all types; it uses the presigned URL with
    `download` attribute so the browser saves the file.
11. The panel contains an empty **Tags** section with a TODO comment marking the Epic 4
    integration point. No analysis is triggered.
12. Pressing `Esc` or a close button dismisses the panel; canvas focus is restored.
13. Opening a different node while the panel is open replaces the content.

**E2E test** (`web/e2e/detail.spec.ts`):

14. After uploading a fixture zip containing at least one image, clicks an image node and
    confirms: the panel opens, the `<img>` is visible, metadata is shown, the download
    link is present.
15. Pressing `Esc` closes the panel.

### Out of Scope

- AI tagging or any analysis call (Epic 4 seam only).
- Full-screen / expanded preview mode.
- Commenting or annotation.

### Implementation Notes

- `GET /api/documents/{id}` can reuse the existing `GetDocument` service method
  (presigned URL already generated there).
- The panel is rendered as a fixed/absolute overlay column so the xyflow canvas underneath
  remains interactive.
- Fresh presigned URL: fetched via `getDocument` on each open (not cached in the node
  state, which may be stale after 15 min).
- The Tags section should be a clearly-named component placeholder (`<DocumentTags />`
  returning `null`) so Epic 4 can slot in without changing the panel structure.

### Files

| File | Action |
|------|--------|
| `internal/adapter/inbound/httpserver/handler/document.go` | Update — GET /{id} handler |
| `internal/adapter/inbound/httpserver/handler/document_test.go` | Update (functional) |
| `internal/adapter/inbound/httpserver/server.go` | Update — register GET /{id} |
| `web/src/documents/api.ts` | Update — `getDocument` |
| `web/src/components/DocumentDetail.tsx` | Create — panel + preview + download + tags seam |
| `web/src/components/DocumentCanvas.tsx` | Update — emit node selection |
| `web/src/pages/Workspace.tsx` | Update — render `DocumentDetail` for selected node |
| `web/src/index.css` | Update — detail panel styles |
| `web/e2e/detail.spec.ts` | Create |

---

## Verification

- `make test` (functional) green: `GET /{id}` metadata + presign URL, ownership 404, 401.
- `make lint` clean.
- `make e2e` green: click node → panel opens → image visible → Esc closes.
- `npx tsc --noEmit` clean.
- Manual: open an image, a PDF, and a text file; confirm each renders or falls back
  correctly. Confirm the download saves the file.
