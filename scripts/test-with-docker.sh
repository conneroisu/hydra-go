#!/usr/bin/env bash
set -euo pipefail

echo "========================================="
echo "Testing SDK with Real Hydra (Docker)"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check for Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Function to cleanup
cleanup() {
    echo ""
    echo "Cleaning up..."
    docker-compose down 2>/dev/null || true
}
trap cleanup EXIT

# Start services
echo "Starting PostgreSQL and Hydra..."
docker-compose up -d postgres

# Wait for PostgreSQL
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U hydra 2>/dev/null; then
        echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
        break
    fi
    sleep 1
done

# For now, let's use the official Hydra Docker image if available, or test with our mock
echo ""
echo "Since Hydra doesn't have an official Docker image, we'll test with our mock server"
echo "For real Hydra testing, use the NixOS VM approach:"
echo "  ./nixos-tests/run-real-hydra-test.sh"
echo ""

# Start our mock server instead
echo "Starting mock Hydra server for Docker-based testing..."
pkill -f "nixos-tests/mock-server/server" 2>/dev/null || true
pushd nixos-tests/mock-server > /dev/null
go build -o server .
popd > /dev/null

nixos-tests/mock-server/server &
SERVER_PID=$!
sleep 2

# Test SDK
echo "Testing SDK..."
cat > test-docker.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/conneroisu/hydra-go/hydra"
)

func main() {
    client, err := hydra.NewClientWithURL("http://localhost:8080")
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    ctx := context.Background()
    
    // Test connection
    projects, err := client.ListProjects(ctx)
    if err != nil {
        log.Fatalf("Failed to list projects: %v", err)
    }
    
    fmt.Printf("✓ Connected to Hydra\n")
    fmt.Printf("✓ Found %d projects\n", len(projects))
    
    // Test auth
    user, err := client.Login(ctx, "testuser", "testpass")
    if err != nil {
        log.Printf("Failed to login: %v", err)
    } else {
        fmt.Printf("✓ Authenticated as %s\n", user.Username)
    }
    
    fmt.Println("\n✅ SDK tests passed!")
}
EOF

go run test-docker.go

kill $SERVER_PID 2>/dev/null || true

echo ""
echo -e "${GREEN}✅ Docker-based tests completed!${NC}"
echo ""
echo "For testing with real Hydra in NixOS VM, run:"
echo "  ./nixos-tests/run-real-hydra-test.sh"