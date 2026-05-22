# Epic 4 — Document Detail View

## Goal

Let a user click a document on the canvas to open a detailed view: a larger preview, its
metadata, and a download action. This view is also where on-demand AI tagging (Epic 5) will
later hook in — so it ships with a clearly-marked, empty seam for tags.

## Why

Closes the MVP loop: upload → see on canvas → **open and inspect**. It also fixes the
integration point for the first analysis feature, so adding AI later is additive, not a
redesign.

## Outcome

- A detail panel/modal that opens on node selection.
- Inline preview for images and PDFs; a download link for other types — all via presigned
  GET URLs.
- A visible-but-empty "Tags" area marked as the future AI seam (no AI built).

## Component stories

- **US-24** — Document detail panel
- **US-25** — Preview & download
- **US-26** — Canvas → detail wiring + AI seam

---

## US-24 — Document detail panel

**As a** user,
**I want** a detail view of a document I clicked,
**So that** I can see what it is and its metadata.

### Acceptance Criteria

1. `DocumentDetail` opens when a canvas node is selected and closes via an explicit control
   (and `Esc`).
2. Shows filename, content type, size (human-readable), and upload date.
3. Renders as a side panel or modal that does not destroy canvas state underneath.

### Files

| File | Action |
|------|--------|
| `web/src/components/DocumentDetail.tsx` | Create |
| `web/src/index.css` | Update — detail panel styles |

---

## US-25 — Preview & download

**As a** user,
**I want** to preview the document and download the original,
**So that** I can read it in place or keep a copy.

### Acceptance Criteria

1. Images render inline (`<img>` from the presigned GET URL).
2. PDFs render in an `<iframe>`/`<embed>`; unsupported types show a type icon + filename.
3. A "Download" action uses the presigned GET URL to fetch the original file.
4. The detail view fetches a fresh presigned URL (via `getDocument`) so links don't expire.

### Files

| File | Action |
|------|--------|
| `web/src/components/DocumentDetail.tsx` | Update — preview + download |

---

## US-26 — Canvas → detail wiring + AI seam

**As a** developer,
**I want** selection state wired from the canvas to the detail view, with a placeholder for
future tagging,
**So that** the loop is complete and Epic 5 has a defined integration point.

### Acceptance Criteria

1. Selecting a node opens its detail; deselecting/closing returns focus to the canvas.
2. The detail view includes a visible **Tags** section that is currently empty, with a clear
   `TODO`/comment marking it as the on-demand AI tagging seam (Epic 5). **No AI is built.**
3. Opening a document does not yet trigger any analysis call (seam only).

### Files

| File | Action |
|------|--------|
| `web/src/components/DocumentCanvas.tsx` | Update — emit selection |
| `web/src/pages/Workspace.tsx` | Update — render `DocumentDetail` for selection |

---

## Verification

- `npx tsc --noEmit` clean.
- In the browser: click an image node → preview + metadata shown → download works; click a
  PDF → renders in-panel; close returns to the canvas. Tags area is present but empty.
- Extend Playwright e2e to cover upload → canvas → open detail (cross-cutting, see OVERVIEW).
