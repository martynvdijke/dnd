#!/usr/bin/env bash
set -e
gofmt -l .
test -z "$(gofmt -l .)"
