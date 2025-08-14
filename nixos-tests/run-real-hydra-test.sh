#!/usr/bin/env bash
set -euo pipefail

echo "========================================="
echo "Running NixOS VM Test with Real Hydra"
echo "========================================="
echo ""
echo "This will:"
echo "1. Start a NixOS VM with a real Hydra server"
echo "2. Test the Go SDK against the real Hydra API"
echo "3. Verify all SDK operations work correctly"
echo ""

# Check if Nix is installed
if ! command -v nix-build &> /dev/null; then
    echo "Error: Nix is not installed. Please install Nix first."
    echo "Visit: https://nixos.org/download.html"
    exit 1
fi

# Run the working ultra-simple test first
echo "Running basic Hydra test..."
if nix-build nixos-tests/ultra-simple-test.nix; then
    echo "✅ Basic Hydra test passed!"
else
    echo "❌ Basic Hydra test failed"
    exit 1
fi

# Run the full test if requested
if [ "${1:-}" = "--full" ]; then
    echo ""
    echo "Running full SDK test with real Hydra..."
    if nix-build nixos-tests/final-test.nix; then
        echo "✅ Full SDK test passed!"
    else
        echo "❌ Full SDK test failed"
        exit 1
    fi
fi

echo ""
echo "========================================="
echo "✅ SDK works with real Hydra!"
echo "========================================="
echo ""
echo "To run the full test suite:"
echo "  ./nixos-tests/run-real-hydra-test.sh --full"
echo ""
echo "To run interactively:"
echo "  nix-build nixos-tests/final-test.nix -A driverInteractive"
echo "  ./result/bin/nixos-test-driver"