#!/usr/bin/env bash
# Build script for dnf_last_update
# Designed by Ifesinachi Osude

set -euo pipefail

VERSION=$(cat ../../VERSION)

echo "Building dnf_last_update version ${VERSION}..."

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w -X main.version=${VERSION}" \
  -o dnf_last_update \
  main.go

echo "Build complete: dnf_last_update"
ls -lh dnf_last_update
