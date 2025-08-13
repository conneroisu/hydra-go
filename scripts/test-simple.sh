#!/usr/bin/env bash
set -euo pipefail

echo "Building and starting mock server..."
pkill -f "nixos-tests/mock-server/server" 2>/dev/null || true
pushd nixos-tests/mock-server > /dev/null
go build -o server .
popd > /dev/null

nixos-tests/mock-server/server &
SERVER_PID=$!
sleep 2

echo "Running unit tests..."
go test ./tests/... -short

echo ""
echo "Running integration tests with mock server..."
export HYDRA_TEST_URL=http://localhost:8080
go test -v -tags=integration ./tests/... 2>&1 | grep -E "^(---|\s+)" | grep -E "(PASS|FAIL)"

kill $SERVER_PID 2>/dev/null || true
echo "Done!"