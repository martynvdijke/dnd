#!/usr/bin/env bash
set -e
# check-concatenated-sql.sh — forbid interpolating client input into SQL in handlers/
#
# Two narrowly-scoped rules over non-test handlers/*.go (tests excluded):
#   A. a dynamic json_extract path built by concatenating the '$."' literal
#      (use jsonPathForField() from handlers/sqlident.go and bind the path)
#   B. a request accessor (c.Query/c.Param/c.DefaultQuery/c.PostForm/c.GetHeader)
#      on the same line as a SQL call that also concatenates with '+' — i.e. a
#      client value spliced into SQL text (bind it with '?' instead)
#
# This is deliberately a source-shape guard, not a SQL parser: it catches the
# regression shape cheaply and does not claim to prove data-flow safety.
# Compile-time constant fragments (column lists, IN-list placeholders) are
# expected and must be allowlisted with a justification.
#
# Allowlist: every entry MUST be "path[:line] # justification".
#   - inline ALLOWLIST array below
#   - optional adjacent scripts/ci/check-concatenated-sql.allowlist (same format)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ALLOWLIST_FILE="$SCRIPT_DIR/check-concatenated-sql.allowlist"

# "path[:line] # why this concatenation is not client input"
ALLOWLIST=(
)

entries=()
for e in "${ALLOWLIST[@]}"; do entries+=("$e"); done
if [ -f "$ALLOWLIST_FILE" ]; then
  while IFS= read -r line || [ -n "$line" ]; do
    t="$(echo "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    [ -z "$t" ] && continue
    case "$t" in \#*) continue;; esac
    entries+=("$t")
  done < "$ALLOWLIST_FILE"
fi

failed=0
patterns=()
for e in "${entries[@]}"; do
  case "$e" in
    *" # "*)
      j="${e#* # }"
      if [ -z "$(echo "$j" | tr -d '[:space:]')" ]; then
        echo "ERROR: allowlist entry missing justification: '$e'" >&2
        failed=1
      fi
      ;;
    *)
      echo "ERROR: allowlist entry missing ' # justification': '$e'" >&2
      failed=1
      ;;
  esac
  patterns+=("$(echo "${e%% #*}" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')")
done
if [ "$failed" -ne 0 ]; then
  exit 1
fi

offenders=""
if [ -d handlers ]; then
  # Rule A: dynamic json_extract path literal.
  offenders="$(grep -rn --include='*.go' -e 'json_extract' handlers/ 2>/dev/null | grep -v '_test\.go' | grep -F '$."' || true)"
  # Rule B: request accessor + concatenation + SQL call on one line.
  rule_b="$(grep -rn --include='*.go' -E 'c\.(Query|Param|DefaultQuery|PostForm|GetHeader)\(' handlers/ 2>/dev/null \
    | grep -v '_test\.go' \
    | grep -E '(db\.DB\.|tx\.)(Query|QueryRow|Exec)' \
    | grep -F '+' || true)"
  if [ -n "$rule_b" ]; then
    offenders="${offenders}${offenders:+$'\n'}${rule_b}"
  fi
fi

if [ -z "$offenders" ]; then
  exit 0
fi

unallowed=""
while IFS= read -r line; do
  [ -z "$line" ] && continue
  file_line="$(echo "$line" | cut -d: -f1,2)"
  file_only="$(echo "$line" | cut -d: -f1)"
  ok=0
  for pat in "${patterns[@]}"; do
    if [ "$file_line" = "$pat" ] || [ "$file_only" = "$pat" ]; then
      ok=1
      break
    fi
  done
  if [ "$ok" -eq 0 ]; then
    unallowed="${unallowed}${line}"$'\n'
  fi
done <<< "$offenders"

if [ -n "$unallowed" ]; then
  echo "ERROR: possible SQL injection — client input concatenated into SQL in handlers/:" >&2
  printf "%s" "$unallowed" >&2
  echo "" >&2
  echo "Fix:" >&2
  echo "  - json_extract paths: use jsonPathForField() from handlers/sqlident.go and bind the path with '?'" >&2
  echo "  - values/limits: bind them with '?' (see handlers/sqlident.go, docs/adr-data-access.md)" >&2
  echo "If a concatenated fragment is a compile-time constant, add a justified allowlist entry:" >&2
  echo "  path/to/file.go[:line] # why this fragment is a constant, not client input" >&2
  exit 1
fi

exit 0
