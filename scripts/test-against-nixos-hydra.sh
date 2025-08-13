#!/usr/bin/env bash
set -euo pipefail

echo "========================================="
echo "Testing SDK Against NixOS Hydra Instance"
echo "========================================="
echo ""
echo "This will test the SDK against the public NixOS Hydra instance"
echo "at https://hydra.nixos.org (read-only operations only)"
echo ""

# Test against public Hydra
export HYDRA_URL="https://hydra.nixos.org"

echo "Building test program..."
go build -o hydra-test nixos-tests/test-real-hydra.go

echo ""
echo "Running tests against $HYDRA_URL..."
echo ""

./hydra-test

rm -f hydra-test

echo ""
echo "To test against a local Hydra instance:"
echo "  1. Run a NixOS VM with Hydra:"
echo "     nix-build nixos-tests/minimal-hydra-test.nix"
echo ""
echo "  2. Or use the full test suite:"
echo "     ./nixos-tests/run-real-hydra-test.sh"