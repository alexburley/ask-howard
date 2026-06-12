#!/usr/bin/env bash
set -euo pipefail

API="${API_URL:-http://localhost:8080}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURE="$SCRIPT_DIR/../web/e2e/fixtures/test-content.zip"
EMAIL="dev@ask-howard.local"
PASSWORD="password123"
COOKIE_JAR="$(mktemp)"

trap 'rm -f "$COOKIE_JAR"' EXIT

json() { python3 -c "import sys,json; print(json.load(sys.stdin)['$1'])"; }

echo "→ Registering dev user ($EMAIL)..."
REG_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST "$API/api/auth/register" \
  -H "Content-Type: application/json" \
  -c "$COOKIE_JAR" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

if [ "$REG_STATUS" != "201" ]; then
  echo "  User already exists — logging in..."
  curl -s -f -X POST "$API/api/auth/login" \
    -H "Content-Type: application/json" \
    -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" > /dev/null
fi

echo "→ Requesting upload slot..."
SLOT=$(curl -s -f -X POST "$API/api/documents/upload" \
  -H "Content-Type: application/json" \
  -b "$COOKIE_JAR" \
  -d '{"filename":"test-content.zip"}')

SET_ID=$(echo "$SLOT" | json document_set_id)
PRESIGNED_URL=$(echo "$SLOT" | json presigned_url)

echo "→ Uploading fixture zip..."
curl -s -f -X PUT "$PRESIGNED_URL" \
  -H "Content-Type: application/zip" \
  --data-binary "@$FIXTURE" > /dev/null

echo "→ Completing upload..."
curl -s -f -X POST "$API/api/documents/sets/$SET_ID/complete" \
  -b "$COOKIE_JAR" > /dev/null

echo "→ Waiting for processing..."
for i in $(seq 1 30); do
  RESULT=$(curl -s -f -b "$COOKIE_JAR" "$API/api/documents/sets/$SET_ID")
  STATUS=$(echo "$RESULT" | json status)
  if [ "$STATUS" = "READY" ]; then
    COUNT=$(echo "$RESULT" | json document_count)
    echo ""
    echo "✓ Seeded — $COUNT documents extracted"
    echo "  Open http://localhost:5173 and log in as $EMAIL / $PASSWORD"
    exit 0
  elif [ "$STATUS" = "FAILED" ]; then
    echo "Processing failed" >&2; exit 1
  fi
  echo "  ($i/30) $STATUS..."
  sleep 2
done

echo "Timed out waiting for processing" >&2
exit 1
