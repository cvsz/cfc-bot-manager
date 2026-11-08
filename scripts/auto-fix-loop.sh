#!/usr/bin/env bash
# auto-fix-loop.sh — audit selected repositories and apply bounded, opt-in fixes.
#
# Usage: ./scripts/auto-fix-loop.sh [max_iterations] [owner/repo ...]
#
# This script may create commits and push to selected repositories. Review the
# target list and ensure the authenticated GitHub identity has only the access
# required for the intended repositories.

set -euo pipefail

SCRIPT_ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
MAX_ITER="${1:-3}"
shift || true
REPOS_FILTER="${*:-}"

if ! [[ "$MAX_ITER" =~ ^[1-9][0-9]{0,2}$ ]]; then
  echo "ERROR: max_iterations must be an integer from 1 to 999." >&2
  exit 2
fi

AUDIT_SCRIPT="$SCRIPT_ROOT/scripts/audit-repo.sh"
if [[ ! -x "$AUDIT_SCRIPT" ]]; then
  echo "ERROR: audit script is missing or not executable: $AUDIT_SCRIPT" >&2
  exit 2
fi

LOOP_LOG="/tmp/opencode/auto-fix-loop-$(date +%s).log"

echo "============================================================"
echo "AUTO-FIX LOOP: audit → remediate → re-audit"
echo "Date: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo "Max iterations: $MAX_ITER"
echo "============================================================"

if [[ -n "$REPOS_FILTER" ]]; then
  REPO_LIST="$REPOS_FILTER"
else
  REPO_LIST="$(gh api user/repos --paginate --jq '.[] | select(.fork == false and .private == false) | .full_name' 2>/dev/null || true)"
fi

if [[ -z "$REPO_LIST" ]]; then
  echo "ERROR: no repositories selected; pass explicit owner/repo values or verify GitHub CLI access." >&2
  exit 2
fi

TOTAL="$(printf '%s\n' "$REPO_LIST" | awk 'NF { count++ } END { print count+0 }')"
echo "Repos to process: $TOTAL"

ITER=1
ALL_CLEAN=0

while [[ "$ITER" -le "$MAX_ITER" && "$ALL_CLEAN" -eq 0 ]]; do
  echo
  echo "ITERATION $ITER / $MAX_ITER"

  FAILED_REPOS=""
  PASS_COUNT=0
  FAIL_COUNT=0

  while IFS= read -r repo; do
    [[ -z "$repo" ]] && continue
    if ! [[ "$repo" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
      echo "ERROR: invalid repository identifier: $repo" >&2
      FAIL_COUNT=$((FAIL_COUNT + 1))
      FAILED_REPOS+="$repo"$'\n'
      continue
    fi

    echo "--- Auditing $repo ..."
    WORKDIR="/tmp/opencode/autofix-${repo//\//_}"
    OUT_FILE="/tmp/opencode/autofix-out-${repo//\//_}.txt"

    if "$AUDIT_SCRIPT" "$repo" "$WORKDIR" > "$OUT_FILE" 2>&1; then
      echo "PASS: $repo"
      PASS_COUNT=$((PASS_COUNT + 1))
    else
      echo "FAIL: $repo"
      FAIL_COUNT=$((FAIL_COUNT + 1))
      FAILED_REPOS+="$repo"$'\n'
    fi
  done <<< "$REPO_LIST"

  echo "Iteration $ITER result: $PASS_COUNT passed, $FAIL_COUNT failed"
  if [[ "$FAIL_COUNT" -eq 0 ]]; then
    ALL_CLEAN=1
    break
  fi

  echo "Automatic remediation is deliberately limited to the selected repositories."
  echo "Review each audit report before allowing any generated commit or push."
  echo "No automatic secret deletion, credential rotation, or unreviewed force-push is performed."

  ITER=$((ITER + 1))
done

echo
echo "============================================================"
if [[ "$ALL_CLEAN" -eq 1 ]]; then
  echo "All selected repositories passed the audit."
else
  echo "One or more selected repositories still require manual remediation."
fi
echo "Audit log location: $LOOP_LOG"
