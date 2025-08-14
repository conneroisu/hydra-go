#!/usr/bin/env bash
set -euo pipefail

echo "========================================="
echo "Hydra Go SDK Test Runner (Testcontainers)"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: go.mod not found. Please run this script from the project root.${NC}"
    exit 1
fi

echo "1. Running unit tests..."
echo "----------------------------"
if go test -v ./tests/... -short; then
    echo -e "${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "${RED}✗ Unit tests failed${NC}"
    exit 1
fi

echo ""
echo "2. Running integration tests with testcontainers..."
echo "----------------------------------------------------"
echo -e "${BLUE}This will automatically start a Hydra Docker container${NC}"
echo -e "${BLUE}Image: ghcr.io/conneroisu/hydra-go/hydra-test:latest${NC}"
echo ""

# Run integration tests with testcontainers
if go test -v -tags=integration ./tests/... 2>&1 | tee test-output.log; then
    echo -e "${GREEN}✓ Integration tests completed${NC}"
else
    echo -e "${RED}✗ Integration tests failed${NC}"
    echo -e "${YELLOW}Check test-output.log for details${NC}"
    exit 1
fi

echo ""
echo "3. Test Summary"
echo "---------------"
PASSES=$(grep -c "PASS:" test-output.log 2>/dev/null || echo "0")
FAILS=$(grep -c "FAIL:" test-output.log 2>/dev/null || echo "0")

echo "Passed: $PASSES"
echo "Failed: $FAILS"

if [ "$FAILS" -eq "0" ]; then
    echo -e "${GREEN}✅ All tests passed with testcontainers!${NC}"
    echo ""
    echo "Benefits of the new approach:"
    echo "• Real Hydra instance (not mock)"
    echo "• Automated container management" 
    echo "• Consistent testing environment"
    echo "• No manual server setup required"
    rm -f test-output.log
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Check test-output.log for details.${NC}"
    exit 1
fi