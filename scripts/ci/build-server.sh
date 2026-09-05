#!/usr/bin/env bash
set -e
VERSION="${VERSION:-$(git describe --tags --always 2>/dev/null || echo 0.0.0-dev)}"
go build -tags sqlite_fts5 -ldflags "-X main.Version=${VERSION}" -o villum-server .
