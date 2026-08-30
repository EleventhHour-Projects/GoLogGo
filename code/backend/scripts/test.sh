#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${BACKEND_DIR}"

if ! command -v go >/dev/null 2>&1; then
  echo "Go toolchain is required but was not found in PATH." >&2
  exit 1
fi

echo "==> Running Go backend tests in ${BACKEND_DIR}"
go test -count=1 ./...
