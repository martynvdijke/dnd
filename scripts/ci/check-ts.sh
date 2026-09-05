#!/usr/bin/env bash
set -e
# Ensure deps are present; use ci if node_modules is missing or stale.
if [ ! -d node_modules ]; then
  npm ci
fi
npm run build:vite
npm run typecheck
