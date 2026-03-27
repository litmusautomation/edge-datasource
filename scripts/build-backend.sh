#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

if command -v go >/dev/null 2>&1; then
  GOPATH_BIN=$(go env GOPATH 2>/dev/null)/bin
  if [[ -d "$GOPATH_BIN" ]]; then
    export PATH="$GOPATH_BIN:$PATH"
  fi
fi

if [[ -d "$HOME/go/bin" ]]; then
  export PATH="$HOME/go/bin:$PATH"
fi

if ! command -v mage >/dev/null 2>&1; then
  printf 'mage is not available in PATH. Install it with `go install github.com/magefile/mage@latest` or ensure your Go bin directory is on PATH.\n' >&2
  exit 1
fi

cd "$ROOT_DIR"
exec mage -v build:linux
