#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT_DIR="$ROOT_DIR/bin"
APP_NAME="explain-html-gen"
ENTRYPOINT="./cmd/explain-html-gen"

mkdir -p "$OUT_DIR"

echo "Building Linux executables..."

for arch in amd64 arm64; do
  output="$OUT_DIR/${APP_NAME}-linux-${arch}"
  echo "  -> $output"
  GOOS=linux GOARCH="$arch" CGO_ENABLED=0 go build -o "$output" "$ENTRYPOINT"
done

echo "Building Windows executables..."

for arch in amd64 arm64; do
  output="$OUT_DIR/${APP_NAME}-windows-${arch}.exe"
  echo "  -> $output"
  GOOS=windows GOARCH="$arch" CGO_ENABLED=0 go build -o "$output" "$ENTRYPOINT"
done

echo "Done. Built artifacts:"
ls -1 "$OUT_DIR"