# Epic 6 — Knowledge Base (Deferred)

> **Status: Deferred.** Outline only — not part of the MVP. Builds on the tags (Epic 4) and
> relationships (Epic 5) to form a queryable picture of a user's document collection.

## Goal

Extract structured **entities** (people, places, dates, events) and **facts** from documents
and assemble them into a per-user knowledge graph that surfaces "interesting facts" and can
answer questions about the collection — "who appears most often?", "what's the timeline?".

## Why

This is the payoff of the whole pipeline: not just storing and relating documents, but
understanding them — turning scattered records into a coherent family narrative.

## Approach (direction)

- Use Claude (Epic 4's analyzer) to extract structured entities/facts with citations back to
  the source document.
- Store entities and relationships in Postgres (normalised tables or a graph extension)
  scoped per user, with provenance (document, passage).
- Reconcile duplicate entities (same person across documents) using tags/embeddings.
- Provide a query/summary surface: timelines, person profiles, surfaced "interesting facts".

## Story outline (full-stack, to be detailed when scheduled)

- **US-xx** — Entity/fact schema with provenance.
- **US-xx** — Extraction job behind an `Extractor` port. Functional test with fixture docs.
- **US-xx** — Entity reconciliation/deduplication. Functional test.
- **US-xx** — Knowledge query API (timeline, person profile, facts feed). Functional test.
- **US-xx** — Knowledge UI (facts feed, person/timeline views, jump-to-source). e2e test.

## Open questions

- Graph store: stay in Postgres (recursive queries) or adopt a dedicated graph layer?
- Automatic extraction vs. user-confirmed (accuracy vs. friction trade-off)?
- Ranking heuristics for "interesting" facts to avoid noise.

## Dependencies

- Epic 4 (analysis/extraction), Epic 5 (entity linking via embeddings), Epic 1 (jobs).
