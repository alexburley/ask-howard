FROM golangci/golangci-lint:latest-alpine AS golangci-lint-bin

FROM golang:1.26-alpine AS ci
COPY --from=golangci-lint-bin /usr/bin/golangci-lint /usr/local/bin/golangci-lint
RUN apk add --no-cache wget postgresql-client && \
    ARCH=$(uname -m) && \
    case "$ARCH" in \
        aarch64) ATLAS_ARCH=arm64 ;; \
        x86_64)  ATLAS_ARCH=amd64 ;; \
    esac && \
    wget -q "https://release.ariga.io/atlas/atlas-community-linux-${ATLAS_ARCH}-latest" \
        -O /usr/local/bin/atlas && \
    chmod +x /usr/local/bin/atlas
RUN go install gotest.tools/gotestsum@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
WORKDIR /app

FROM golang:1.26-alpine AS dev
RUN go install github.com/air-verse/air@latest
WORKDIR /app
EXPOSE 8080
CMD ["air"]

FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM golang:1.26-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /ask-howard ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=go-builder /ask-howard /ask-howard
EXPOSE 8080
ENTRYPOINT ["/ask-howard"]
