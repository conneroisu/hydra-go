#!/usr/bin/env bash
set -euo pipefail

echo "========================================="
echo "Hydra Go SDK Test Verification"
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

echo "1. Running Go unit tests..."
echo "----------------------------"
if go test -v ./hydra/... -short; then
    echo -e "${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "${RED}✗ Unit tests failed${NC}"
    exit 1
fi

echo ""
echo "2. Running integration tests..."
echo "--------------------------------"
# Start mock server in background for local testing
echo "Starting mock server..."
cd nixos-tests/mock-server
go run . &
SERVER_PID=$!
cd ../..

# Wait for server to start
sleep 2

# Check if server is running
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Mock server is running${NC}"
else
    echo -e "${RED}✗ Mock server failed to start${NC}"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

# Run integration tests
export HYDRA_TEST_URL=http://localhost:8080
if go test -v -tags=integration ./tests/...; then
    echo -e "${GREEN}✓ Integration tests passed${NC}"
else
    echo -e "${RED}✗ Integration tests failed${NC}"
    kill $SERVER_PID
    exit 1
fi

# Stop mock server
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null || true

echo ""
echo "3. Running test coverage analysis..."
echo "-------------------------------------"
go test -race -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo "Total coverage: $COVERAGE"

# Check if coverage meets threshold (e.g., 70%)
COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
if (( $(echo "$COVERAGE_NUM > 70" | bc -l) )); then
    echo -e "${GREEN}✓ Coverage meets threshold (>70%)${NC}"
else
    echo -e "${YELLOW}⚠ Coverage below threshold (<70%)${NC}"
fi

echo ""
echo "4. Running linting checks..."
echo "-----------------------------"
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run ./...; then
        echo -e "${GREEN}✓ Linting passed${NC}"
    else
        echo -e "${YELLOW}⚠ Linting issues found${NC}"
    fi
else
    echo -e "${YELLOW}⚠ golangci-lint not installed, skipping${NC}"
fi

echo ""
echo "5. Checking for race conditions..."
echo "-----------------------------------"
if go test -race ./hydra/...; then
    echo -e "${GREEN}✓ No race conditions detected${NC}"
else
    echo -e "${RED}✗ Race conditions detected${NC}"
    exit 1
fi

echo ""
echo "6. Running benchmarks..."
echo "------------------------"
go test -bench=. -benchmem ./hydra/... -run=^$ | head -20
echo -e "${GREEN}✓ Benchmarks completed${NC}"

echo ""
echo "========================================="
echo -e "${GREEN}All tests passed successfully!${NC}"
echo "========================================="
echo ""
echo "Test Summary:"
echo "- Unit tests: ✓"
echo "- Integration tests: ✓"
echo "- Coverage: $COVERAGE"
echo "- Race detection: ✓"
echo ""
echo "To run NixOS VM tests, execute:"
echo "  ./nixos-tests/run-vm-test.sh"
echo ""
echo "To generate HTML coverage report:"
echo "  go tool cover -html=coverage.out -o coverage.html"
echo "  open coverage.html"