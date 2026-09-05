#!/usr/bin/env bash
set -e
go test -tags "sqlite_fts5" -v -coverprofile=coverage.out -covermode=atomic -coverpkg=./handlers/...,./middleware/... ./...
go tool cover -func=coverage.out | tee /dev/stderr | awk '
  /^total:/ {
    gsub(/%/, "", $3)
    if ($3 + 0 < 20) {
      printf "FAIL: total coverage %.1f%% < 20%% threshold\n", $3
      exit 1
    }
    printf "PASS: total coverage %.1f%% >= 20%% threshold\n", $3
  }'
awk '/^mode:/{print; next} $1 ~ /^villum\/handlers\//{print}' coverage.out > /tmp/handlers.out
awk '/^mode:/{print; next} $1 ~ /^villum\/middleware\//{print}' coverage.out > /tmp/middleware.out
H=$(go tool cover -func=/tmp/handlers.out | awk '/^total:/ { gsub(/%/, "", $3); printf "%.1f", $3 }')
M=$(go tool cover -func=/tmp/middleware.out | awk '/^total:/ { gsub(/%/, "", $3); printf "%.1f", $3 }')
printf "handlers/ coverage: %s%% (threshold 40%%)\n" "$H"
printf "middleware/ coverage: %s%% (threshold 40%%)\n" "$M"
awk -v h="$H" -v m="$M" 'BEGIN {
  fail = 0
  if (h + 0 < 40) { print "FAIL: handlers/ coverage below 40% threshold"; fail = 1 }
  if (m + 0 < 40) { print "FAIL: middleware/ coverage below 40% threshold"; fail = 1 }
  exit fail
}'
go tool cover -func=coverage.out | awk '/^total:/ { gsub(/%/, "", $3); printf "%.1f\n", $3 }' > /tmp/cov-total.txt
