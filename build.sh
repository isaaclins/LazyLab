#!/usr/bin/env bash
set -euo pipefail

# Build lazylab CLI into ./bin
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

mkdir -p bin
APP_NAME="lazylab"
GOOS_ENV=${GOOS:-}
GOARCH_ENV=${GOARCH:-}
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION=${VERSION:-dev}
LDFLAGS=${LDFLAGS:-"-s -w -X github.com/isaaclins/LazyLab/cmd.Version=$VERSION -X github.com/isaaclins/LazyLab/cmd.Commit=$GIT_COMMIT -X github.com/isaaclins/LazyLab/cmd.Date=$BUILD_DATE"}

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
