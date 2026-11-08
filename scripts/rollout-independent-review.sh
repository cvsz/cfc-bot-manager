#!/usr/bin/env bash
# Roll out the independent-review workflow through reviewable pull requests.
# Requires: gh, git, jq, and authenticated access to the target repositories.
set -euo pipefail

OWNER="${OWNER:-cvsz}"
REVIEWER="${REVIEWER:-}"
APPLY="${APPLY:-0}"
WORKFLOW=".github/workflows/request-independent-review.yml"
BRANCH="chore/independent-review-governance"
SCRIPT_ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_WORKFLOW="$SCRIPT_ROOT/$WORKFLOW"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

if ! [[ "$OWNER" =~ ^[A-Za-z0-9-]+$ ]]; then
  echo "ERROR: OWNER must be a GitHub user or organization login." >&2
  exit 2
fi
if ! [[ "$REVIEWER" =~ ^[A-Za-z0-9-]+$ ]]; then
  echo "ERROR: Set REVIEWER to a valid, independent GitHub login." >&2
  exit 2
fi
if [[ "${REVIEWER,,}" == "${OWNER,,}" ]]; then
  echo "ERROR: REVIEWER must differ from OWNER." >&2
  exit 2
fi
if [[ "$APPLY" != "0" && "$APPLY" != "1" ]]; then
  echo "ERROR: APPLY must be 0 (dry-run) or 1 (create reviewable PRs)." >&2
  exit 2
fi
if [[ ! -f "$SOURCE_WORKFLOW" ]]; then
  echo "ERROR: Canonical workflow not found: $SOURCE_WORKFLOW" >&2
  exit 2
fi

owner_type="$(gh api "users/$OWNER" --jq '.type')"
case "$owner_type" in
  User)
    repos_endpoint="users/$OWNER/repos?per_page=100&type=owner"
    ;;
  Organization)
    repos_endpoint="orgs/$OWNER/repos?per_page=100&type=all"
    ;;
  *)
    echo "ERROR: Unsupported GitHub owner type: $owner_type" >&2
    exit 2
    ;;
esac

gh api --paginate "$repos_endpoint" \
  --jq '.[] | [.full_name, .default_branch, .archived, .fork, .disabled, (.mirror_url // "")] | @tsv' |
while IFS=$'\t' read -r repo base archived fork disabled mirror_url; do
  if [[ "$archived" == true || "$fork" == true || "$disabled" == true || -n "$mirror_url" ]]; then
    printf 'SKIP %s (archived/fork/disabled/mirror)\n' "$repo"
    continue
  fi
  if ! gh api "repos/$repo/collaborators/$REVIEWER/permission" --jq '.permission' 2>/dev/null | grep -Eq '^(write|maintain|admin)$'; then
    printf 'BLOCKED %s (reviewer lacks write access)\n' "$repo"
    continue
  fi
  if gh api "repos/$repo/contents/$WORKFLOW?ref=$base" >/dev/null 2>&1; then
    printf 'SKIP %s (workflow already exists; review manually)\n' "$repo"
    continue
  fi
  if gh api "repos/$repo/branches/$BRANCH" >/dev/null 2>&1; then
    printf 'SKIP %s (rollout branch already exists)\n' "$repo"
    continue
  fi

  printf 'READY %s base=%s reviewer=%s\n' "$repo" "$base" "$REVIEWER"
  [[ "$APPLY" == "1" ]] || continue

  repo_dir="$TMP/repo"
  rm -rf "$repo_dir"
  mkdir -p "$repo_dir"
  gh repo clone "$repo" "$repo_dir" -- --depth 1 --branch "$base"
  (
    cd "$repo_dir"
    git checkout -q -b "$BRANCH"
    mkdir -p .github/workflows
    cp "$SOURCE_WORKFLOW" "$WORKFLOW"
    git add "$WORKFLOW"
    git commit -qm "ci: establish independent PR review gate"
    git push -q -u origin "$BRANCH"
  )
  gh pr create -R "$repo" --head "$BRANCH" --base "$base" \
    --title "ci: establish independent PR review gate" \
    --body "Adds a request workflow and exact-head independent-approval check. Before merge, configure repository variable INDEPENDENT_REVIEWER=$REVIEWER and require the verify-independent-approval status check. Does not approve, merge, or bypass branch rules."
  printf 'PR CREATED %s\n' "$repo"
done

echo "NOTE: A requested review is not an enforced gate unless verify-independent-approval is required by branch protection."
