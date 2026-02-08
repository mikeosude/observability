#!/usr/bin/env bash
# Build script for dnf_update_check
# Designed by Ifesinachi Osude

set -euo pipefail

VERSION=$(cat ../../VERSION)

echo "Building dnf_update_check version ${VERSION}..."

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w -X main.version=${VERSION}" \
  -o dnf_update_check \
  main.go

echo "Build complete: dnf_update_check"
ls -lh dnf_update_check
