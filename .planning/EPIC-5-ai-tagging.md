# Epic 5 — On-demand AI Tagging (Deferred)

> **Status: Deferred.** Outline only — not part of the MVP. Decisions are captured here so
> the work is ready to pick up. Epic 4 (US-26) ships the UI seam this epic plugs into.

## Goal

When a user opens a document, the system analyzes it on demand and applies descriptive
**tags** (e.g. "census record", "photograph", "marriage certificate", surnames, places,
dates). Tags accumulate into a per-user vocabulary that can later filter/group the canvas.

## Why

Tagging is the first layer of intelligence and the foundation for relationships (Epic 6) and
the knowledge base (Epic 7). On-demand keeps cost proportional to attention — we only pay to
analyze what users actually open.

## Approach (decided)

- **Claude API** (multimodal) for tagging and light fact extraction.
- **OCR step** for scanned images/PDFs (Tesseract sidecar or a hosted OCR API) so text-bearing
  scans become analyzable text before the Claude call.
- An outbound **`Analyzer` port** keeps the provider swappable; the concrete adapter wraps
  `anthropic-sdk-go`.
- Analysis runs as a **River job** (reuse Epic 2's queue) so opening a document enqueues
  work and the UI polls/streams results into the Tags area.

## Story outline

- **US-2x** — `Analyzer` outbound port + Claude adapter (multimodal call, structured tag output).
- **US-2x** — OCR step for scanned images/PDFs feeding the analyzer.
- **US-2x** — `tags` + per-user `tag_vocabulary` schema and repository.
- **US-2x** — Tagging job + enqueue-on-open; idempotent re-runs.
- **US-2x** — Detail-view Tags UI (replace the Epic 4 seam): show, poll, and let users
  confirm/remove tags.
- **US-2x** — Canvas filter/group by tag.

## Open questions

- Tag schema: free-form strings vs. a controlled taxonomy (people/place/date/doc-type)?
- Do user edits to tags feed back into prompts/quality?
- Cost controls: per-user quotas, caching by content hash to avoid re-analysis.

## Dependencies

- Epic 2 (River queue, documents) and Epic 4 (detail-view seam).
