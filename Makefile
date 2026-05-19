.PHONY: start dev build test test-unit lint fmt kill web-install web-dev web-build infra-up infra-down docker-build migrate-diff migrate-apply migrate-status

# Start everything: infrastructure, Go hot-reload server, and Vite dev server.
# Requires: air (go install github.com/air-verse/air@latest) and hivemind (brew install hivemind)
start:
	docker compose up -d --wait --remove-orphans
	hivemind

# Run only the API server in dev mode (no hot reload, no frontend).
dev:
	go run -tags dev ./cmd/server

# Production build: compiles frontend then embeds it into the Go binary.
build: web-build
	go build -o bin/ask-howard ./cmd/server

test:
	go test -tags functional ./...

test-unit:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	golangci-lint fmt ./...

kill:
	@lsof -ti :8080,:5173 | xargs kill -9 2>/dev/null || true

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

# Start local infrastructure (Postgres). Required before running tests.
infra-up:
	docker compose up -d --remove-orphans

infra-down:
	docker compose down

docker-build:
	docker build -t ask-howard:local .

# Usage: make migrate-diff name=describe_your_change
migrate-diff:
	atlas migrate diff --config internal/adapter/outbound/postgres/atlas.hcl --env local "$(name)"

migrate-apply:
	atlas migrate apply --config internal/adapter/outbound/postgres/atlas.hcl --env local

migrate-status:
	atlas migrate status --config internal/adapter/outbound/postgres/atlas.hcl --env local
