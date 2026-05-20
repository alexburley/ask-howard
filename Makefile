.PHONY: start clean-start build test t coverage test-unit lint fmt generate sqlc migrate-diff migrate-apply migrate-status e2e

# Run a command in the ci container without starting postgres.
CI = docker compose run --rm --no-deps --build ci
# Run a command in the ci container, starting postgres first (via depends_on).
CI_DEPS = docker compose run --rm --build ci

# Start everything in Docker: Postgres, Go API (air hot-reload), and Vite dev server.
start:
	docker compose up --build --remove-orphans

# Tear down volumes and start fresh.
clean-start:
	docker compose down -v
	$(MAKE) start

# Build the production Docker image.
build:
	docker build -t ask-howard:local .

# Run functional tests (requires Docker socket for testcontainers).
test:
	docker compose run --rm --no-deps --build \
		-v /var/run/docker.sock:/var/run/docker.sock \
		ci go test -tags functional ./...

t:
	docker compose run --rm --no-deps --build \
		-v /var/run/docker.sock:/var/run/docker.sock \
		ci sh -c 'gotestsum --format testname -- -tags functional -coverprofile=coverage.out -covermode=atomic ./... && go tool cover -func=coverage.out | tail -1'

coverage:
	docker compose run --rm --no-deps --build \
		-v /var/run/docker.sock:/var/run/docker.sock \
		ci sh -c 'gotestsum --format testname -- -tags functional -coverprofile=coverage.out -covermode=atomic ./... && go tool cover -func=coverage.out'

test-unit:
	$(CI) gotestsum --format testname ./...

lint:
	$(CI) golangci-lint run ./...

fmt:
	$(CI) golangci-lint fmt ./...

generate: sqlc

sqlc:
	$(CI) sqlc generate

# Usage: make migrate-diff name=describe_your_change
migrate-diff:
	docker compose run --rm --build \
		-v /var/run/docker.sock:/var/run/docker.sock \
		ci atlas migrate diff \
		--config file://internal/adapter/outbound/postgres/atlas.hcl \
		--env local \
		--var "database_url=postgres://ask-howard:ask-howard@postgres:5432/ask-howard?sslmode=disable" \
		"$(name)"

migrate-apply:
	$(CI_DEPS) atlas migrate apply \
		--config file://internal/adapter/outbound/postgres/atlas.hcl \
		--env local \
		--var "database_url=postgres://ask-howard:ask-howard@postgres:5432/ask-howard?sslmode=disable"

# Run Playwright e2e tests. Requires `make start` to be running first.
e2e:
	docker compose --profile e2e run --rm playwright

migrate-status:
	$(CI_DEPS) atlas migrate status \
		--config file://internal/adapter/outbound/postgres/atlas.hcl \
		--env local \
		--var "database_url=postgres://ask-howard:ask-howard@postgres:5432/ask-howard?sslmode=disable"
