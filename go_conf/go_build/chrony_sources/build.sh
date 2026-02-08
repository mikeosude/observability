#!/usr/bin/env bash
# Build script for chrony_sources
# Designed by Ifesinachi Osude

set -euo pipefail

VERSION=$(cat ../../VERSION)

echo "Building chrony_sources version ${VERSION}..."

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w -X main.version=${VERSION}" \
  -o chrony_sources \
  main.go

echo "Build complete: chrony_sources"
ls -lh chrony_sources
