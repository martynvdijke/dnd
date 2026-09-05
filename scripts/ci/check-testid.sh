#!/usr/bin/env bash
set -e
FAIL=0
while read -r tid; do
  [ -z "$tid" ] && continue
  if ! grep -rq -- "$tid" tests/; then
    # nav-* testids are constructed dynamically in helpers.ts (nav-${nav});
    # accept them when the nav id appears in tests alongside the template
    case "$tid" in
      nav-*)
        suffix="${tid#nav-}"
        # shellcheck disable=SC2016
        if grep -rq 'nav-\${nav}' tests/helpers.ts && grep -rq -- "$suffix" tests/; then
          continue
        fi
        ;;
    esac
    echo "data-testid \"$tid\" has no test reference"
    FAIL=1
  fi
done < <(grep -rhoE 'data-testid="[^"]+"' ts/ static/app.html static/login.html static/admin.html | sed 's/.*="\(.*\)"/\1/' | sort -u)
exit $FAIL
