#!/usr/bin/env bash
set -e
# Verbatim CI gate: plain build with no -ldflags version override, so the
# binary reports the default 0.0.0-dev version (the smoke e2e asserts the
# footer matches /v\d+\.\d+\.\d+/; a describe-derived short SHA would break it
# in tag-less checkouts). Release/local versioned builds live in Taskfile.
go build -tags sqlite_fts5 -o villum-server .
