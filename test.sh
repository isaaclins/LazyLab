#!/usr/bin/env bash
set -euo pipefail

# Run tests for lazylab
# Usage:
#   ./test.sh                 # run unit/integration tests with race+cover
#   E2E=1 ./test.sh           # additionally attempt E2E (opt-in, requires Docker; currently informational)
#   VERBOSE=1 ./test.sh       # pass -v to go test
#   PKG=./profiles ./test.sh  # test specific package (defaults to ./...)

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

PKG=${PKG:-./...}
VERBOSE_FLAG=""
[[ "${VERBOSE:-0}" == "1" ]] && VERBOSE_FLAG="-v"

# Unit/Integration tests (table-driven, non-docker)
echo "==> Running Go tests ($PKG)"
go test ${VERBOSE_FLAG} -race -cover "$PKG"

# Optional E2E (Docker) - interactive flows not suitable for CI by default
if [[ "${E2E:-0}" == "1" ]]; then
  if command -v docker >/dev/null 2>&1; then
    echo "==> E2E note: lazylab is interactive by design. For a quick manual check, run:"
    echo "    ./build.sh && ./bin/lazylab --no-net --graceful --purge --image alpine:3.20 --shell sh"
    echo "    # then 'exit' to end."
  else
    echo "==> Skipping E2E: docker not available"
  fi
fi

echo "==> Tests completed"
