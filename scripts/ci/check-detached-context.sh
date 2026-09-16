#!/usr/bin/env bash
set -e
# check-detached-context.sh — forbid context.TODO()/Background() in request-scoped handlers/
#
# Fails if context.TODO() or context.Background() appears in any non-test .go
# file under handlers/. Tests ( *_test.go ) are excluded.
#
# Allowlist: genuinely detached work that legitimately outlives any request
# (e.g. startup tasks, background jobs) may use Background/TODO if documented.
# Two forms are supported:
#   1. Inline array ALLOWLIST below — each entry is "path[:line] # justification"
#   2. Adjacent file scripts/ci/check-detached-context.allowlist — same line format
# Every allowlist entry MUST include a one-line justification after " # ".
# The script fails if any allowlist entry is missing a justification.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ALLOWLIST_FILE="$SCRIPT_DIR/check-detached-context.allowlist"

# Inline allowlist — add entries as "path[:line] # why this is genuinely detached".
# Example: "handlers/jobs.go:42 # background cleanup job outlives request"
ALLOWLIST=(
)

FAIL=0

# Collect allowlist entries from inline array and optional file.
allowlist_entries=()
for e in "${ALLOWLIST[@]}"; do
  allowlist_entries+=("$e")
done
if [ -f "$ALLOWLIST_FILE" ]; then
  while IFS= read -r line || [ -n "$line" ]; do
    # Skip empty lines and comment-only lines
    trimmed="$(echo "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    [ -z "$trimmed" ] && continue
    case "$trimmed" in \#*) continue;; esac
    allowlist_entries+=("$trimmed")
  done < "$ALLOWLIST_FILE"
fi

# Validate every allowlist entry has a justification comment.
for entry in "${allowlist_entries[@]}"; do
  case "$entry" in
    *" # "*)
      # Ensure justification is non-empty after " # "
      justification="${entry#* # }"
      # trim whitespace
      justification_trimmed="$(echo "$justification" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
      if [ -z "$justification_trimmed" ]; then
        echo "ERROR: allowlist entry missing justification: '$entry'" >&2
        echo "  Each allowlist entry must be 'path[:line] # justification of why detached'" >&2
        FAIL=1
      fi
      ;;
    *)
      echo "ERROR: allowlist entry missing ' # justification': '$entry'" >&2
      echo "  Each allowlist entry must be 'path[:line] # justification of why detached'" >&2
      FAIL=1
      ;;
  esac
done
if [ "$FAIL" -ne 0 ]; then
  exit 1
fi

# Build a set of allowed paths/patterns (strip justification for matching).
allowed_patterns=()
for entry in "${allowlist_entries[@]}"; do
  pat="${entry%% #*}"
  pat="$(echo "$pat" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
  allowed_patterns+=("$pat")
done

# Find offenders in handlers/ excluding _test.go
offenders=""
# Use grep -rn; filter out _test.go lines
if [ -d "handlers" ]; then
  offenders="$(grep -rn --include='*.go' -E 'context\.(TODO|Background)\(\)' handlers/ 2>/dev/null | grep -v '_test\.go' || true)"
fi

if [ -z "$offenders" ]; then
  exit 0
fi

# Check each offender against allowlist
unallowed=""
while IFS= read -r line; do
  [ -z "$line" ] && continue
  # line is like "handlers/foo.go:123:  code"
  file_line="$(echo "$line" | cut -d: -f1,2)"
  file_only="$(echo "$line" | cut -d: -f1)"
  allowed=0
  for pat in "${allowed_patterns[@]}"; do
    if [ "$file_line" = "$pat" ] || [ "$file_only" = "$pat" ]; then
      allowed=1
      break
    fi
  done
  if [ "$allowed" -eq 0 ]; then
    unallowed="${unallowed}${line}\n"
  fi
done <<< "$offenders"

if [ -n "$unallowed" ]; then
  echo "ERROR: detached context usage found in handlers/ (use c.Request.Context() instead):" >&2
  printf "%b" "$unallowed" >&2
  echo "" >&2
  echo "Fix: replace context.Background()/context.TODO() with c.Request.Context() (or a derived context)" >&2
  echo "for request-scoped work so cancellation/deadlines propagate." >&2
  echo "" >&2
  echo "If this is genuinely detached work that outlives any request (background job," >&2
  echo "startup task — see spec \"Background context remains valid for genuinely detached work\")," >&2
  echo "add a justified allowlist entry to one of:" >&2
  echo "  - scripts/ci/check-detached-context.sh ALLOWLIST array" >&2
  echo "  - scripts/ci/check-detached-context.allowlist" >&2
  echo "Format: path/to/file.go[:line] # justification of why this context legitimately outlives the request" >&2
  echo "Example: handlers/jobs.go:42 # background cleanup job outlives request" >&2
  exit 1
fi

exit 0
