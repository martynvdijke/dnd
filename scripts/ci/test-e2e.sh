#!/usr/bin/env bash
set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
"$SCRIPT_DIR/build-server.sh"
rm -f villum.db villum.db-wal villum.db-shm
npx playwright test --project=chromium
