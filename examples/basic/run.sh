#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

REPO_ROOT="$(cd ../.. && pwd)"
BIN_DIR="$(pwd)/bin"

mkdir -p "$BIN_DIR"

echo "Building saga and couriers..."
go build -o "$BIN_DIR/saga" "$REPO_ROOT"
go build -o "$BIN_DIR/saga-courier-slack-legacy" "$REPO_ROOT/couriers/saga-courier-slack-legacy"
go build -o "$BIN_DIR/saga-courier-basecamp-messageboard" "$REPO_ROOT/couriers/saga-courier-basecamp-messageboard"

export PATH="$BIN_DIR:$PATH"
export SAGA_COURIER_SLACK_LEGACY__WEBHOOK_URL="https://hooks.slack.com/services/T/B/X"
export SAGA_COURIER_BASECAMP_MESSAGEBOARD__REFRESH_TOKEN="demo-token"
export SAGA_COURIER_BASECAMP_MESSAGEBOARD__CLIENT_ID="demo-client"
export SAGA_COURIER_BASECAMP_MESSAGEBOARD__CLIENT_SECRET="demo-secret"
export SAGA_COURIER_BASECAMP_MESSAGEBOARD__ACCOUNT_ID="9999999"

echo
echo "=== saga tell 0.1.0 --dry-run ==="
saga --dir .saga tell 0.1.0 --dry-run

echo
echo "=== saga tell 0.1.0 ==="
saga --dir .saga tell 0.1.0
