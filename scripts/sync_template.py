#!/usr/bin/env python3
"""Safely propose allowlisted upstream template updates without overwriting local edits."""

import argparse
import json
import os
import subprocess
import sys
import urllib.error
import urllib.request
from pathlib import Path

ORIGIN = "cvsz/ztemplate"
INITIAL_BASELINE = "a19231153425488c307ce2009d74aad79fce0e8b"
STATE = Path(".github/template-sync-state.json")
MANAGED = (
    "docs/ai/prompts/universal-autonomous-engineering.md",
    "docs/ai/prompts/repository.md",
    "docs/ai/playbooks/autonomous-repo-upgrade.md",
)


def get_text(ref, path):
    url = f"https://raw.githubusercontent.com/{ORIGIN}/{ref}/{path}"
    request = urllib.request.Request(url, headers={"User-Agent": "zeaz-template-sync"})
    with urllib.request.urlopen(request, timeout=30) as response:
        return response.read().decode("utf-8")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--source-ref", default="main")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()
    if not Path("scripts/validate_repo.py").is_file():
        parser.error("Run from the repository root")
    state = json.loads(STATE.read_text(encoding="utf-8")) if STATE.exists() else {}
    previous = state.get("source_ref", INITIAL_BASELINE)
    try:
        resolved = subprocess.check_output(
            ["git", "ls-remote", f"https://github.com/{ORIGIN}.git", f"refs/heads/{args.source_ref}"],
            text=True, timeout=30,
        ).split()[0]
    except (subprocess.CalledProcessError, subprocess.TimeoutExpired, IndexError) as exc:
        print(f"ERROR: Cannot resolve template source ref: {exc}", file=sys.stderr)
        return 2
    changed = []
    skipped = []
    for filename in MANAGED:
        path = Path(filename)
        # Avoid following local symlinks or modifying paths outside the repository.
        if path.is_symlink() or any(parent.is_symlink() for parent in path.parents):
            skipped.append(f"{filename}: symlink path")
            continue
        try:
            latest = get_text(resolved, filename)
            baseline = get_text(previous, filename)
        except (urllib.error.URLError, UnicodeError) as exc:
            print(f"ERROR: Cannot retrieve upstream {filename}: {exc}", file=sys.stderr)
            return 2
        local = path.read_text(encoding="utf-8") if path.exists() else None
        if local == latest:
            continue
        if local is not None and local != baseline:
            skipped.append(f"{filename}: locally modified; manual reconciliation required")
            continue
        changed.append((path, latest))
    for item in skipped:
        print(f"SKIP: {item}")
    # Do not advance the tracked baseline if any managed files need manual reconciliation.
    if skipped:
        print("BLOCKED: resolve local conflicts; no automatic changes made", file=sys.stderr)
        return 3
    if args.dry_run:
        for path, _ in changed:
            print(f"WOULD UPDATE: {path}")
        return 0
    for path, content in changed:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")
        print(f"UPDATED: {path}")
    if changed or previous != resolved:
        STATE.parent.mkdir(parents=True, exist_ok=True)
        STATE.write_text(json.dumps({"source_ref": resolved}, indent=2) + "\n", encoding="utf-8")
    print(f"Template sync: {len(changed)} files updated")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
