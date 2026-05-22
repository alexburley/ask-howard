# Epic 4 — On-demand AI Tagging (Deferred)

> **Status: Deferred.** Outline only — not part of the MVP. Decisions are captured here so
> the work is ready to pick up. Epic 3 (US-15) ships the `<DocumentTags />` seam this epic
> fills.

## Goal

When a user opens a document, the system analyses it on demand and applies descriptive
**tags** (e.g. "census record", "photograph", "marriage certificate", surnames, places,
dates). Tags accumulate into a per-user vocabulary that can later filter/group the canvas.

## Why

Tagging is the first layer of intelligence and the foundation for relationships (Epic 5) and
the knowledge base (Epic 6). On-demand keeps cost proportional to attention — we only pay to
analyse what users actually open.

## Approach (decided)

- **Claude API** (multimodal) for tagging and light fact extraction.
- **OCR step** for scanned images/PDFs (Tesseract sidecar or a hosted OCR API) so
  text-bearing scans become analysable text before the Claude call.
- An outbound **`Analyzer` port** keeps the provider swappable; the concrete adapter wraps
  `anthropic-sdk-go`.
- Analysis runs as a **River job** (reuse Epic 1's queue) so opening a document enqueues
  work and the UI polls/streams results into the Tags panel.

## Story outline (full-stack, to be detailed when scheduled)

- **US-xx** — `Analyzer` outbound port + Claude adapter: multimodal call, structured tag
  output. Functional test with a fixture image.
- **US-xx** — OCR step feeding the analyzer for scanned images/PDFs.
- **US-xx** — `tags` + `tag_vocabulary` schema, repository, and tag-apply service method.
- **US-xx** — Tagging job enqueued on document open; idempotent re-runs; polling endpoint.
- **US-xx** — Detail panel Tags UI (replaces `<DocumentTags />` seam from US-15): shows
  tags, polling state, and user confirm/remove actions. e2e test.
- **US-xx** — Canvas filter/group by tag. e2e test.

## Open questions

- Tag schema: free-form strings vs. a controlled taxonomy (person/place/date/doc-type)?
- Do user edits feed back into prompts to improve future quality?
- Cost controls: per-user quotas, caching by content hash to avoid re-analysis.

## Dependencies

- Epic 1 (River queue, documents) and Epic 3 (detail panel seam).
