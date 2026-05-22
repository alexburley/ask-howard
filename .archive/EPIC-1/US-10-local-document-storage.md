# US-10 — Local document storage ✓ DONE

**Completed:** 2026-05-22
**Commit:** `1acce9a`

## User Story

**As a** developer,
**I want** a working S3-compatible object store in the local stack,
**So that** the upload and extraction features have durable blob storage to build on.

## What was delivered

- `docker-compose.yml`: `minio` service (API `:9000`, console `:9001`) + `minio-init` bucket bootstrap + `minio_data` volume. `api` receives `S3_*` env vars and depends on `minio-init`; `ci` receives `TEST_S3_*` env vars.
- `internal/port/outbound/objectstore.go`: `ObjectStore` interface (`PresignPut`, `PresignGet`, `Get`, `Put`, `Delete`).
- `internal/adapter/outbound/s3/store.go`: `aws-sdk-go-v2` adapter with path-style support for MinIO. `Config` passed by pointer. `var _ outbound.ObjectStore = (*Store)(nil)` compile-time assertion.
- `internal/adapter/outbound/s3/store_test.go`: functional tests (`//go:build functional`) — `Put`/`Get`/`Delete` round-trip and presigned PUT+GET flow. Skip when `TEST_S3_ENDPOINT` unset.
- `cmd/server/main.go`: S3 config in `loadConfig`; store constructed in `run()` helper (extracted to fix `exitAfterDefer` lint issue).

## Notes

- `NewStore` takes `*Config` (not `Config`) to satisfy `hugeParam` lint rule.
- `run()` helper pattern adopted to cleanly separate defer cleanup from `os.Exit`.
- `_ = objectStore` placeholder — wired to document service in US-11.
