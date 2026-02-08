#!/usr/bin/env bash
# Build all RHEL stats collectors
# Designed by Ifesinachi Osude

set -euo pipefail

VERSION=$(cat VERSION)
echo "=========================================="
echo "Building RHEL Stats Collectors v${VERSION}"
echo "Designed by Ifesinachi Osude"
echo "=========================================="
echo ""

cd go_build/dnf_last_update
./build.sh
echo ""

cd ../dnf_update_check
./build.sh
echo ""

cd ../chrony_sources
./build.sh
echo ""

cd ../..
echo "=========================================="
echo "All builds complete!"
echo "=========================================="
echo ""
echo "Binaries created:"
ls -lh go_build/*/dnf_* go_build/*/chrony_*
