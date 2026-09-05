#!/usr/bin/env bash
set -e
if ! docker info >/dev/null 2>&1; then
  echo "WARNING: Docker daemon not available — skipping Docker build smoke test"
  exit 0
fi
docker build . --file Dockerfile --tag dnd-smoke:test
