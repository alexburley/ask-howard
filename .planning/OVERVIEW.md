# Ask Howard — Document Platform Overview

## Vision

Ask Howard lets people upload **ancestry documents** (as zip archives, often large and
image-heavy), explore them on an interactive **canvas** where each document floats and can
be opened for a detailed view, and — over time — have the system **analyze, tag, relate, and
extract facts** from them into a queryable knowledge base.

This document is the top-level planning artifact. It captures the technical direction, the
key decisions, and the epic map. Each epic has its own file (`EPIC-N-*.md`) breaking the
work into **full-stack user stories** — each story delivers a working slice of user-facing
functionality, covering backend, frontend, functional tests, and e2e tests together.

## Where we are today

The codebase is early-stage:

- **Backend (Go, hexagonal):** auth (register / login / logout / me via JWT cookies),
  a health check, a Postgres `users` table. No object storage, upload, jobs, or AI yet.
- **Frontend (React 19 + Vite):** auth pages, an auth context, a health banner, and a
  near-empty `Workspace` page.
- **Infra:** Docker Compose (Postgres, Go API with air, Vite). Atlas migrations, sqlc,
  Playwright e2e, GitHub Actions CI.

Existing planning stories `US-5`–`US-9` cover the auth/app-shell work already built. New
stories continue numbering from **US-10**.

## Decisions (confirmed with product owner)

| Topic | Decision |
|---|---|
| AI approach (deferred epics) | **Claude API + OCR** for tagging/extraction; embeddings for relationship/similarity search |
| Document ownership | **Per-user private sets** — all document data scoped by `user_id` |
| Upload / storage | **Presigned direct-to-S3 upload** of the zip + **server-side extraction** into per-document objects |
| Background jobs | **River** (Postgres-backed queue) — durable, pgx-native, no new infra |
| Canvas | **`@xyflow/react`** (React Flow) — grows into the relationship-graph epic without a rewrite |
| Story orientation | **Full-stack vertical slices** — each story delivers frontend + backend + functional test + e2e |

## Technical overview

### Stack additions

- **Object storage:** MinIO locally (added to Compose), Scaleway Object Storage in prod.
  Client: `aws-sdk-go-v2` (`s3` + presign) — S3-compatible, supports presigned URLs.
- **Background jobs:** [River](https://github.com/riverqueue/river) (`riverqueue/river` +
  `riverpgxv5`) — durable, retries, survives restarts; its own migration adds River's tables.
- **Canvas:** `@xyflow/react` — pan / zoom / drag out of the box; nodes now, relationship
  edges later.
- **Deferred AI:** `anthropic-sdk-go` (Claude, multimodal) for tagging/extraction; an OCR
  step (Tesseract sidecar or hosted OCR) for scanned images/PDFs; `pgvector` for embeddings.

### Why presigned upload + server-side extraction

Proxying multi-GB zips through the Go API is the thing we'd later have to rip out. Instead:

1. Client requests an upload slot → API returns a **presigned PUT URL** + a `document_set`
   row (`status = UPLOADING`).
2. Client uploads the zip **directly to S3**.
3. Client calls **complete** → API marks `PROCESSING` and enqueues a **River extraction job**.
4. Worker downloads the zip, streams entries, writes each document as its own S3 object,
   inserts one `documents` row each, sets the set to `READY` (or `FAILED` + error).

### Hexagonal placement (follows existing patterns)

```
internal/
├── domain/                         DocumentSet, Document, sentinel errors
├── port/inbound/                   DocumentService
├── port/outbound/                  DocumentRepository, ObjectStore, JobEnqueuer
├── service/                        document.go (application service)
└── adapter/
    ├── inbound/httpserver/handler/ document.go (HTTP endpoints)
    └── outbound/
        ├── postgres/               document.go repo + sqlc queries
        ├── s3/                     ObjectStore adapter
        └── jobs/                   River client + extraction worker
```

**Auth note:** token parsing is currently inline in each handler. New authenticated endpoints
should use a shared `handler.currentUserID(r, jwtSecret)` helper returning a 401
`problem.DetailedError` on failure.

## Test strategy

Every story that changes the backend ships a **functional test** (`//go:build functional`,
runs with `make test` against the compose Postgres and MinIO). Every story that changes the
frontend ships a **Playwright e2e test** in `web/e2e/`. Both must be green before a story is
considered done.

## Epic map

| Epic | Title | Status | Stories |
|---|---|---|---|
| 1 | Document upload | MVP | US-10 .. US-12 |
| 2 | Document canvas | MVP | US-13 .. US-14 |
| 3 | Document detail | MVP | US-15 |
| 4 | On-demand AI tagging | Deferred | outline only |
| 5 | Relationship analysis | Deferred | outline only |
| 6 | Knowledge base | Deferred | outline only |

**MVP = epics 1–3:** a user can upload a zip, watch it extract, see their documents floating
on a canvas, rearrange them (positions persist), and click one to view/download it.

## Sequencing

1. **US-10** (storage infra) is the foundation — nothing else works without it.
2. **US-11** and **US-12** deliver the upload pipeline end-to-end; US-12 depends on US-11.
3. **US-13** (canvas) and **US-14** (layout) are frontend-led but need US-12's document list
   API; they can be built in parallel with each other after US-12's API lands.
4. **US-15** (detail) builds on US-13's selection model and US-12's presigned GET.
5. Epics 4–6 are deferred; US-15 ships a clearly-marked seam for AI tagging (Epic 4).
