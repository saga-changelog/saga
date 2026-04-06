#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

REPO_ROOT="$(cd ../.. && pwd)"
BIN_DIR="$(pwd)/bin"
SAGA_DIR=".saga"

echo "Reset state: move any chaptered feats back to pending..."
if [ -d "$SAGA_DIR/feats/chapters" ]; then
  for dir in "$SAGA_DIR"/feats/chapters/*/; do
    [ -d "$dir" ] && mv "$dir"*.jsonnet "$SAGA_DIR/feats/pending/" 2>/dev/null || true
    rmdir "$dir" 2>/dev/null || true
  done
fi

mkdir -p "$BIN_DIR"

echo "Building saga..."
go build -o "$BIN_DIR/saga" "$REPO_ROOT"

echo "Building saga-courier-stdout..."
go build -o "$BIN_DIR/saga-courier-stdout" "$REPO_ROOT/couriers/saga-courier-stdout"

export PATH="$BIN_DIR:$PATH"
export SAGA_COURIER_STDOUT__NOTE="hello from the stdout courier"
SAGA="saga --dir $SAGA_DIR"

echo
echo "=== validate ==="
$SAGA validate

echo
echo "=== pending ==="
$SAGA pending

echo
echo "=== chapters create 1.0.0 ==="
$SAGA chapters create 1.0.0

echo
echo "=== tell 1.0.0 --dry-run ==="
$SAGA tell 1.0.0 --dry-run

echo
echo "=== tell 1.0.0 ==="
$SAGA tell 1.0.0
