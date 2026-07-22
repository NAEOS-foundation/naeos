#!/usr/bin/env bash
set -euo pipefail

echo "=== NAEOS E2E Test Suite ==="
echo ""

echo "--- Running integration-tagged E2E tests (pkg/pipeline) ---"
go test -tags=integration -race -count=1 -timeout 120s ./pkg/pipeline/ -run "TestEndToEnd" -v

echo ""
echo "--- Running full test suite with race detector ---"
go test -race -count=1 -timeout 300s ./...

echo ""
echo "--- Building CLI ---"
go build -o /dev/null ./cmd/naeos/

echo ""
echo "=== All E2E tests passed ==="
