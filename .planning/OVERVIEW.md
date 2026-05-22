# Ask Howard — Document Platform Overview

## Vision

Ask Howard lets people upload **ancestry documents** (as zip archives, often large and
image-heavy), explore them on an interactive **canvas** where each document floats and can
be opened for a detailed view, and — over time — have the system **analyze, tag, relate, and
extract facts** from them into a queryable knowledge base.

This document is the top-level planning artifact. It captures the technical direction, the
key decisions, and the epic map. Each epic has its own file (`EPIC-N-*.md`) breaking the
work into component user stories that follow the existing `US-N` convention in this folder.

## Where we are today

The codebase is early-stage:

- **Backend (Go, hexagonal):** auth (register / login / logout / me via JWT cookies),
  a health check, a Postgres `users` table. No object storage, upload, jobs, or AI yet.
- **Frontend (React 19 + Vite):** auth pages, an auth context, a health banner, and a
  near-empty `Workspace` page.
- **Infra:** Docker Compose (Postgres, Go API with air, Vite). Atlas migrations, sqlc,
  Playwright e2e, GitHub Actions CI.

Existing planning stories `US-5`–`US-9` cover the auth/app-shell work already built. New
stories in these epics continue numbering from **US-10**.

## Decisions (confirmed with product owner)

| Topic | Decision |
|---|---|
| AI approach (deferred epics) | **Claude API + OCR** for tagging/extraction; embeddings for relationship/similarity search |
| Document ownership | **Per-user private sets** — all document data scoped by `user_id` |
| Upload / storage | **Presigned direct-to-S3 upload** of the zip + **server-side extraction** into per-document objects |
| Background jobs | **River** (Postgres-backed queue) — durable, pgx-native, no new infra |
| Canvas | **`@xyflow/react`** (React Flow) — grows into the relationship-graph epic without a rewrite |
| Scope | MVP (epics 1–4) detailed to story level; analysis epics (5–7) outlined and deferred |

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

This handles large files cleanly, and the per-document objects/rows are exactly what the
canvas, detail view, and later analysis need. Tradeoff: requires bucket CORS config and a
small client-side PUT flow.

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

**Auth note:** there is no auth middleware today — `/auth/me` parses the JWT cookie inline
(`token.Parse`). New authenticated endpoints should reuse a small extracted helper, e.g.
`handler.currentUserID(r, jwtSecret) (uuid.UUID, error)`, returning a 401
`problem.DetailedError` on failure.

## Epic map

| Epic | Title | Status | Stories |
|---|---|---|---|
| 1 | Object storage adapter | MVP | US-10 .. US-12 |
| 2 | Upload pipeline | MVP | US-13 .. US-18 |
| 3 | Canvas workspace | MVP | US-19 .. US-23 |
| 4 | Document detail view | MVP | US-24 .. US-26 |
| 5 | On-demand AI tagging | Deferred | outline only |
| 6 | Relationship analysis | Deferred | outline only |
| 7 | Knowledge base | Deferred | outline only |

**MVP = epics 1–4:** a user can upload a zip, watch it process, see their documents floating
on a canvas, rearrange them (positions persist), and click one to view/download it. No AI yet.

## Suggested sequencing

1. **Epic 1** (storage) is the foundation — nothing uploads without it.
2. **Epic 2** (pipeline) depends on Epic 1; delivers the data model + extraction.
3. **Epic 3** (canvas) and **Epic 4** (detail) are frontend-led and depend on Epic 2's API.
   They can be built in parallel once the document endpoints exist.
4. **Epics 5–7** are deferred; Epic 4 leaves a clearly-marked seam where on-demand tagging
   (Epic 5) will hook in.

## Cross-cutting (applies across MVP epics)

- Update `README.md` and `CLAUDE.md` for object storage, jobs, upload flow, and new env vars.
- Add `http/documents.http` manual request file.
- Extend Playwright e2e (`web/e2e/`): register → upload a small fixture zip → wait for
  processing → see nodes on canvas → open detail. Unique email per run.
- Per repo rules: `make test` (functional) and `make lint` green before any story is "done".
