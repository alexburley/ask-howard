# Epic 1 — Object Storage Adapter

## Goal

Give the backend an S3-compatible object store it can use for all blob storage: original
zip uploads and the individual documents extracted from them. MinIO runs locally via
Compose; Scaleway Object Storage is the production target. Both are reached through one
outbound `ObjectStore` port so the rest of the app never imports an S3 SDK.

## Why

Everything downstream — uploads, the canvas, the detail view, and later analysis — needs
durable blob storage with **presigned URLs** (so large files move client↔S3 directly,
never through the Go process). This epic establishes that foundation.

## Outcome

- `ObjectStore` outbound port: `PresignPut`, `PresignGet`, `Get`, `Put`, `Delete`.
- An `aws-sdk-go-v2` adapter that works against MinIO (path-style) and Scaleway/AWS.
- MinIO + bucket bootstrap in Docker Compose, with config plumbed through `main.go`.
- A functional test proving round-trip and presign behaviour against MinIO.

## Component stories

- **US-10** — MinIO in the local stack
- **US-11** — `ObjectStore` port + S3 adapter
- **US-12** — Object store functional test

---

## US-10 — MinIO in the local stack

**As a** developer,
**I want** an S3-compatible object store running in the local Compose stack with a
pre-created bucket,
**So that** the API and tests can store and retrieve blobs without external infrastructure.

### Acceptance Criteria

1. `docker-compose.yml` defines a `minio` service (API `:9000`, console `:9001`) with a
   named volume and a healthcheck.
2. A one-shot `minio-init` service creates the `ask-howard-docs` bucket on startup
   (idempotent).
3. The `api` service receives S3 env vars and depends on `minio-init` completing.
4. The `ci` service receives `TEST_S3_*` env vars and depends on `minio-init` completing.
5. `make start` brings MinIO up; the console is reachable and the bucket exists.

### Implementation Notes

- Use `minio/minio` for the server and `minio/mc` for the init container
  (`mc mb --ignore-existing local/ask-howard-docs`).
- Env vars: `S3_ENDPOINT`, `S3_BUCKET`, `S3_REGION`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`,
  `S3_USE_PATH_STYLE`. Mirror as `TEST_S3_*` for the `ci` service.
- Add a `minio_data` named volume.

### Files

| File | Action |
|------|--------|
| `docker-compose.yml` | Update — `minio`, `minio-init` services; env on `api`/`ci`; `minio_data` volume |

---

## US-11 — `ObjectStore` port + S3 adapter

**As a** backend developer,
**I want** a single outbound port for object storage with an S3-compatible implementation,
**So that** services depend on an interface and the storage vendor stays swappable.

### Acceptance Criteria

1. `internal/port/outbound/objectstore.go` defines `ObjectStore` with `PresignPut`,
   `PresignGet`, `Get`, `Put`, `Delete`.
2. `internal/adapter/outbound/s3/store.go` implements it with `aws-sdk-go-v2`, asserting
   `var _ outbound.ObjectStore = (*Store)(nil)`.
3. The adapter supports path-style addressing (required by MinIO) and a custom endpoint.
4. Config is read in `cmd/server/main.go` (`loadConfig`) into an `s3.Config`, and a `Store`
   is constructed and wired for use by the document service (Epic 2).
5. `go build ./...` passes.

### Implementation Notes

- `PresignPut(ctx, key, contentType, expiry)` / `PresignGet(ctx, key, expiry)` return URLs
  via `s3.NewPresignClient`.
- `Put(ctx, key, r, size, contentType)` sets `ContentLength` and `ContentType`.
- Construct with `LoadDefaultConfig` + `StaticCredentialsProvider`; set
  `o.BaseEndpoint` and `o.UsePathStyle` in the `s3.NewFromConfig` options func.
- New dep: `github.com/aws/aws-sdk-go-v2/{config,credentials,service/s3}`.

### Files

| File | Action |
|------|--------|
| `internal/port/outbound/objectstore.go` | Create |
| `internal/adapter/outbound/s3/store.go` | Create |
| `cmd/server/main.go` | Update — `s3.Config` in `loadConfig`, construct `Store` |
| `go.mod` / `go.sum` | Update — aws-sdk-go-v2 |

---

## US-12 — Object store functional test

**As a** maintainer,
**I want** a functional test that exercises the real adapter against MinIO,
**So that** storage behaviour (including presigned URLs) is verified, not mocked.

### Acceptance Criteria

1. `internal/adapter/outbound/s3/store_test.go` is tagged `//go:build functional`.
2. The test skips when `TEST_S3_ENDPOINT` is unset (so unit runs are unaffected).
3. Covers: `Put` → `Get` → `Delete` round-trip; presigned PUT (HTTP) then presigned GET
   returns the same bytes.
4. `make test` (functional, with MinIO up) passes.

### Implementation Notes

- Per `CLAUDE.md`: adapters get **functional tests, not stub unit tests**.
- Build the `s3.Config` from `TEST_S3_*` env; use unique keys (`uuid`) per test.
- Use a plain `http.Client` to PUT to the presigned URL with the matching `Content-Type`.

### Files

| File | Action |
|------|--------|
| `internal/adapter/outbound/s3/store_test.go` | Create |

---

## Verification

- `make start` → MinIO console at `http://localhost:9001`, `ask-howard-docs` bucket present.
- `make test` (functional) → S3 round-trip and presign tests green.
- `make lint` → clean.
