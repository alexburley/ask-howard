# Epic 5 — Relationship Analysis (Deferred)

> **Status: Deferred.** Outline only — not part of the MVP. The canvas uses `@xyflow/react`
> specifically so relationships can be drawn as edges without a rewrite.

## Goal

In the background, work out which documents are **related** (same people, places, events,
or time period) and surface those relationships as connecting **edges** on the canvas and as
a "related documents" list in the detail panel.

## Why

Relationships are what turn a pile of files into a story — helping a user discover that a
photo, a census entry, and a certificate all concern the same ancestor.

## Approach (direction)

- Enable the **`pgvector`** extension; store per-document embeddings.
- Build embeddings from extracted text (Epic 4 OCR/analysis output) and/or Claude-generated
  captions for images.
- Compute relatedness via vector nearest-neighbour search, optionally boosted by shared
  structured tags/entities (Epics 4/6) for explainability.
- Run embedding + similarity as **River jobs** triggered after tagging.
- Render relationships as **xyflow edges** on the canvas (toggled, thresholded to keep the
  graph legible).

## Story outline (full-stack, to be detailed when scheduled)

- **US-xx** — `pgvector` migration + embeddings storage.
- **US-xx** — Embedding job behind an outbound `Embedder` port. Functional test.
- **US-xx** — Similarity query + relationship-scoring algorithm (vector + shared-entity
  boost). Functional test.
- **US-xx** — Surface edges on the canvas (toggle on/off, threshold slider). e2e test.
- **US-xx** — "Related documents" list in the detail panel. e2e test.

## Open questions

- Edge threshold and max edges per node to keep the canvas legible.
- Explainability: show *why* two documents are related (shared surname? same year?).
- Embedding model choice, dimensionality, and re-embedding strategy.

## Dependencies

- Epic 4 (text/caption signal), Epic 1 (jobs/documents), Epic 2 (canvas edges).
