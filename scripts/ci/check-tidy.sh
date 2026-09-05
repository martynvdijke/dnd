#!/usr/bin/env bash
set -e
go mod tidy
git diff --exit-code go.mod go.sum
