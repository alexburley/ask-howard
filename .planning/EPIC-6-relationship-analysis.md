# Epic 6 — Relationship Analysis (Deferred)

> **Status: Deferred.** Outline only — not part of the MVP. The canvas was built on
> `@xyflow/react` specifically so relationships can be drawn as edges without a rewrite.

## Goal

In the background, work out which documents are **related** (same people, places, events, or
time period) and surface those relationships — first as connecting **edges** on the canvas,
then as ranked "related documents" in the detail view.

## Why

Relationships are what turn a pile of files into a story. This is the differentiating feature:
helping a user discover that a photo, a census entry, and a certificate all concern the same
ancestor.

## Approach (direction)

- Enable the **`pgvector`** extension; store embeddings per document.
- Build embeddings from extracted text (Epic 5 OCR/analysis output) and/or Claude-generated
  captions/summaries for images.
- Compute relatedness via vector nearest-neighbour search, optionally combined with shared
  structured tags/entities (Epic 5 / Epic 7) for explainable links.
- Run embedding + similarity as **River jobs** triggered after extraction/tagging.
- Render relationships as **xyflow edges**; show "related documents" in the detail view.

## Story outline

- **US-3x** — `pgvector` migration + embeddings storage.
- **US-3x** — Embedding job (text + image-caption) behind an outbound `Embedder` port.
- **US-3x** — Similarity query + a relationship-scoring algorithm (vector + shared-entity boost).
- **US-3x** — Surface edges on the canvas (toggle, thresholded).
- **US-3x** — "Related documents" list in the detail view.

## Open questions

- Edge threshold / max edges per node to keep the canvas legible.
- Explainability: show *why* two documents are related (shared surname? same year?).
- Embedding model choice and dimensionality; re-embedding strategy when analysis improves.

## Dependencies

- Epic 5 (text/caption signal), Epic 2 (jobs/documents), Epic 3 (canvas edges).
