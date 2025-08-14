#!/usr/bin/env bash
set -euo pipefail

# Script to run NixOS VM tests for Hydra Go SDK

echo "========================================="
echo "Running NixOS VM tests for Hydra Go SDK"
echo "========================================="
echo ""

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "Error: go.mod not found. Please run this script from the project root."
    exit 1
fi

# Check if nix is available
if ! command -v nix-build &> /dev/null; then
    echo "Error: nix-build not found. Please install Nix first."
    echo "Visit: https://nixos.org/download.html"
    exit 1
fi

echo "Starting NixOS VM test with real Hydra server..."
echo ""

# Run the test
if nix-build nixos-tests/final-test.nix; then
    echo ""
    echo "========================================="
    echo "✅ SUCCESS: All NixOS VM tests passed!"
    echo "========================================="
    echo ""
    echo "The SDK has been verified to work with:"
    echo "  • Real Hydra server (not mocks)"
    echo "  • PostgreSQL database"
    echo "  • Full NixOS VM environment"
    echo ""
else
    echo ""
    echo "========================================="
    echo "❌ FAILED: NixOS VM tests failed!"
    echo "========================================="
    echo ""
    echo "Troubleshooting:"
    echo "  1. Try the simple test first:"
    echo "     nix-build nixos-tests/ultra-simple-test.nix"
    echo ""
    echo "  2. Check if Hydra alone works:"
    echo "     nix-build '<nixpkgs/nixos/tests/hydra.nix>'"
    echo ""
    echo "  3. View detailed logs:"
    echo "     nix log \$(nix-build nixos-tests/final-test.nix 2>&1 | grep -oE '/nix/store/[^-]+-.*\.drv' | head -1)"
    echo ""
    exit 1
fi

echo "Additional test options:"
echo "  • Basic Hydra test: nix-build nixos-tests/ultra-simple-test.nix"
echo "  • Test against public Hydra: ./test-against-nixos-hydra.sh"
echo ""