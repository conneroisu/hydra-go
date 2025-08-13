#!/usr/bin/env bash
set -euo pipefail

echo "========================================="
echo "Hydra Go SDK Test Runner"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: go.mod not found. Please run this script from the project root.${NC}"
    exit 1
fi

# Kill any existing mock servers
echo "Cleaning up any existing mock servers..."
pkill -f "nixos-tests/mock-server/server" 2>/dev/null || true
sleep 1

echo "1. Building mock server..."
echo "----------------------------"
pushd nixos-tests/mock-server > /dev/null
if go build -o server .; then
    echo -e "${GREEN}✓ Mock server built successfully${NC}"
else
    echo -e "${RED}✗ Failed to build mock server${NC}"
    exit 1
fi
popd > /dev/null

echo ""
echo "2. Running unit tests..."
echo "----------------------------"
if go test -v ./tests/... -short; then
    echo -e "${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "${RED}✗ Unit tests failed${NC}"
    exit 1
fi

echo ""
echo "3. Starting mock server for integration tests..."
echo "-------------------------------------------------"
nixos-tests/mock-server/server &
SERVER_PID=$!
echo "Mock server started with PID: $SERVER_PID"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "Cleaning up..."
    if [ -n "${SERVER_PID:-}" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
}
trap cleanup EXIT

# Wait for server to start
sleep 2

# Check if server is running
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Mock server is running${NC}"
else
    echo -e "${RED}✗ Mock server failed to start${NC}"
    exit 1
fi

echo ""
echo "4. Running integration tests..."
echo "--------------------------------"
export HYDRA_TEST_URL=http://localhost:8080

# Run integration tests
if go test -v -tags=integration ./tests/... 2>&1 | tee test-output.log | grep -E "^(PASS|FAIL|ok|---)" ; then
    echo -e "${GREEN}✓ Integration tests completed${NC}"
else
    echo -e "${YELLOW}⚠ Some integration tests may have issues${NC}"
fi

echo ""
echo "5. Test Summary"
echo "---------------"
PASSES=$(grep -c "PASS:" test-output.log 2>/dev/null || echo "0")
FAILS=$(grep -c "FAIL:" test-output.log 2>/dev/null || echo "0")

echo "Passed: $PASSES"
echo "Failed: $FAILS"

if [ "$FAILS" -eq "0" ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    rm -f test-output.log
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Check test-output.log for details.${NC}"
    exit 1
fi