.PHONY: start dev build test lint web-install web-dev web-build infra-up infra-down docker-build

# Start everything: infrastructure, Go hot-reload server, and Vite dev server.
# Requires: air (go install github.com/air-verse/air@latest) and hivemind (brew install hivemind)
start:
	docker compose up -d --wait
	hivemind

# Run only the API server in dev mode (no hot reload, no frontend).
dev:
	go run -tags dev ./cmd/server

# Production build: compiles frontend then embeds it into the Go binary.
build: web-build
	go build -o bin/pulse ./cmd/server

test:
	go test ./...

lint:
	go vet ./...

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

# Start local infrastructure (Postgres + MinIO). Required before running tests.
infra-up:
	docker compose up -d

infra-down:
	docker compose down

docker-build:
	docker build -t pulse:local .
