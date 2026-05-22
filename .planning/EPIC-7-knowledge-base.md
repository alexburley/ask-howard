# Epic 7 — Knowledge Base (Deferred)

> **Status: Deferred.** Outline only — not part of the MVP. Builds on the tags (Epic 5) and
> relationships (Epic 6) to form a queryable picture of a user's document set.

## Goal

Extract structured **entities** (people, places, dates, events) and **facts** from documents
and assemble them into a per-user knowledge graph that surfaces "interesting facts" and can
answer questions about the collection ("who appears most often?", "what's the timeline?").

## Why

This is the payoff of the whole pipeline: not just storing and relating documents, but
understanding them — turning scattered records into a coherent family narrative.

## Approach (direction)

- Use Claude (Epic 5's analyzer) to extract structured entities/facts with citations back to
  the source document.
- Store entities and their relationships as a graph (normalized Postgres tables, or a graph
  extension) keyed per user, with provenance (which document, which passage).
- Reconcile duplicate entities (same person across documents) using tags/embeddings.
- Provide a query/summary surface: timelines, person profiles, surfaced "interesting facts".

## Story outline

- **US-4x** — Entity/fact schema with provenance.
- **US-4x** — Extraction job producing entities + facts behind an `Extractor` port.
- **US-4x** — Entity reconciliation/deduplication.
- **US-4x** — Knowledge query API (timeline, person profile, facts feed).
- **US-4x** — Knowledge UI (facts feed, person/timeline views, jump-to-source).

## Open questions

- Graph store: stay in Postgres (recursive queries) or adopt a dedicated graph layer?
- How much extraction is automatic vs. user-confirmed (accuracy vs. friction)?
- Surfacing "interesting" facts: ranking heuristics and avoiding noise.

## Dependencies

- Epic 5 (analysis/extraction), Epic 6 (entity linking via embeddings), Epic 2 (jobs).
