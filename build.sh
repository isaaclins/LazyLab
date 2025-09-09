#!/usr/bin/env bash
set -euo pipefail

# Build lazylab CLI into ./bin
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

mkdir -p bin
APP_NAME="lazylab"
GOOS_ENV=${GOOS:-}
GOARCH_ENV=${GOARCH:-}
LDFLAGS=${LDFLAGS:-"-s -w"}

# Determine output extension for Windows
TARGET_OS=${GOOS_ENV:-$(go env GOOS)}
EXT=""
if [[ "$TARGET_OS" == "windows" ]]; then
  EXT=".exe"
fi

# Build
if [[ -n "$GOOS_ENV" || -n "$GOARCH_ENV" ]]; then
  GOOS="${GOOS_ENV:-$(go env GOOS)}" \
  GOARCH="${GOARCH_ENV:-$(go env GOARCH)}" \
  go build -trimpath -ldflags "$LDFLAGS" -o "bin/${APP_NAME}${EXT}" ./main.go
else
  go build -trimpath -ldflags "$LDFLAGS" -o "bin/${APP_NAME}${EXT}" ./main.go
fi

echo "Built $(pwd)/bin/${APP_NAME}${EXT}"
