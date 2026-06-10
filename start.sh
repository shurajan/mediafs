#!/usr/bin/env bash
# Build and start MediaFS on macOS or Linux.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="$SCRIPT_DIR/mediafs"

if ! command -v go >/dev/null 2>&1; then
  echo "❌ Go is not installed. Install it from https://go.dev/dl/" >&2
  exit 1
fi

echo "🔨 Building MediaFS..."
go build -C "$SCRIPT_DIR" -o "$BIN" ./cmd/mediafs

if [ ! -f "$HOME/.mediafs/.meta/auth.json" ]; then
  echo "⚠️  No auth.json found. Create one first:"
  echo "    $BIN hash-password --password <your-password>"
  exit 1
fi

echo "🚀 Starting MediaFS..."
exec "$BIN" "$@"
