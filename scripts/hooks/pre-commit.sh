#!/bin/sh
# pre-commit wrapper for prek (https://github.com/j178/prek)
#
# Why this exists
# ---------------
# prek (like pre-commit) stashes only *unstaged tracked* changes while hooks
# run, leaving untracked files in the working tree. In this repo that breaks
# whole-tree hooks (go vet, go build, tsc, vitest): in-flight work is often
# untracked (e.g. handlers/trmnl.go) while the symbols it depends on live in
# *modified* tracked files that prek has already stashed away. The hooks then
# fail even though the staged commit is perfectly fine.
#
# This wrapper also stashes untracked files (--include-untracked), so hooks
# validate exactly HEAD + index — the contents that will actually be
# committed. A genuinely broken staged state still fails the hooks; only the
# unrelated working-tree WIP is hidden from them.
#
# Install
# -------
#   cp scripts/hooks/pre-commit.sh .git/hooks/pre-commit
#   chmod +x .git/hooks/pre-commit
#
# NOTE: re-running `prek install` overwrites the shim in .git/hooks — copy
# the wrapper again afterwards. Everything here is restored on exit, even if
# hooks fail, so a failed commit never leaves your working tree stashed.

set -u

repo_root="$(git rev-parse --show-toplevel)" || exit 1
hooks_dir="$(git rev-parse --git-path hooks)" || exit 1

PREK="${PREK:-/root/.cargo/bin/prek}"
[ -x "$PREK" ] || PREK="prek"

# Stash unstaged tracked changes AND untracked files, keeping the index.
stashed=0
if git stash push --keep-index --include-untracked --quiet \
    -m "prek-precommit-$(date +%s)-$$" 2>/dev/null; then
    stashed=1
fi

restore() {
    if [ "$stashed" -eq 1 ]; then
        if ! git stash pop --quiet 2>/dev/null; then
            echo "pre-commit: WARNING: could not restore stashed working tree changes" >&2
            echo "pre-commit: see 'git stash list' and run 'git stash pop'" >&2
        fi
    fi
}
trap restore EXIT

"$PREK" hook-impl --hook-dir "$hooks_dir" --script-version 4 \
    --hook-type=pre-commit -- "$@"
rc=$?
exit "$rc"
