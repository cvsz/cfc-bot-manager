"""Regression tests: retired privileged automation must stay retired.

These tests encode the P0 acceptance criteria from Issue #5:
- Unsafe automation is absent from the tree.
- No workflow distributes reviewer credentials (POLIEDBC_PAT) or GPG signing
  secrets (GPG_PASSPHRASE / GPG_SIGNING_KEY).
- No script performs cross-repository mutation or direct-push automation while
  carrying reviewer credentials.
- The deliberate independent-review gate (request-independent-review.yml) is
  preserved and never names the reviewer account in source (it reads an
  Actions variable instead).

These tests inspect the repository tree only; they never execute retired
automation and never read or emit credential values.
"""

from __future__ import annotations

import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
WORKFLOWS = ROOT / ".github" / "workflows"
SCRIPTS = ROOT / "scripts"

# Files retired by the Issue #5 P0 remediation. They must never be reintroduced.
RETIRED_FILES = [
    ".github/workflows/auto-fix-open-prs.yml",
    ".github/workflows/automated-audit-review-approve.yml",
    ".github/scripts/gpg-loopback.sh",
    ".github/scripts/post-review-comment.sh",
    "scripts/auto-fix-loop.sh",
    "scripts/deploy-workflow-all-repos.sh",
    "docs/ai/playbooks/auto-audit-review-approve-policedbc.md",
]

# Secrets/credentials that no committed workflow or script may distribute or
# consume for automated privileged execution.
CREDENTIAL_TOKENS = ("POLIEDBC_PAT", "GPG_PASSPHRASE", "GPG_SIGNING_KEY")

# Markers of cross-repository direct-push / privileged-execution automation.
DIRECT_PUSH_MARKERS = ("gh secret set", "git push origin", "persist-credentials: true")


def _read(rel: str) -> str:
    return (ROOT / rel).read_text(encoding="utf-8")


def _workflow_texts() -> dict[str, str]:
    texts: dict[str, str] = {}
    if WORKFLOWS.is_dir():
        for path in sorted(WORKFLOWS.glob("*.yml")):
            texts[path.name] = path.read_text(encoding="utf-8")
    return texts


class RetiredPrivilegedAutomationTests(unittest.TestCase):
    def test_retired_files_are_absent(self):
        for rel in RETIRED_FILES:
            with self.subTest(file=rel):
                self.assertFalse(
                    (ROOT / rel).exists(),
                    f"retired privileged automation file must not exist: {rel}",
                )

    def test_no_workflow_distributes_reviewer_or_signing_credentials(self):
        texts = _workflow_texts()
        for name, text in texts.items():
            with self.subTest(workflow=name):
                for token in CREDENTIAL_TOKENS:
                    self.assertNotIn(
                        token,
                        text,
                        f"{name} references the retired credential {token}",
                    )

    def test_no_workflow_performs_automated_review_or_approval(self):
        texts = _workflow_texts()
        for name, text in texts.items():
            with self.subTest(workflow=name):
                # Automated reviewer accounts must not be used to post reviews,
                # request changes, or comment through a service-account token.
                self.assertNotIn(
                    "gh pr review",
                    text,
                    f"{name} must not run automated gh pr review",
                )
                self.assertNotIn(
                    "gh pr comment",
                    text,
                    f"{name} must not run automated gh pr comment",
                )
                # No service-account token may be exported as GH_TOKEN for
                # automated privileged execution.
                self.assertNotIn(
                    "GH_TOKEN: ${{ secrets.",
                    text,
                    f"{name} must not expose a repository secret as GH_TOKEN",
                )

    def test_no_script_performs_cross_repo_direct_push_automation(self):
        if not SCRIPTS.is_dir():
            self.skipTest("scripts directory not present")
        for path in sorted(SCRIPTS.rglob("*.sh")):
            text = path.read_text(encoding="utf-8")
            with self.subTest(script=path.relative_to(ROOT).as_posix()):
                # A script that both reaches a credential store and pushes to a
                # remote is cross-repository privileged automation.
                has_credential_access = (
                    "git credential fill" in text or "gh secret set" in text
                )
                has_direct_push = "git push origin" in text
                self.assertFalse(
                    has_credential_access and has_direct_push,
                    f"{path} combines credential access with direct push",
                )

    def test_independent_review_workflow_is_preserved_and_account_agnostic(self):
        workflow = (WORKFLOWS / "request-independent-review.yml").read_text(
            encoding="utf-8"
        )
        self.assertIn("verify-independent-approval:", workflow)
        self.assertIn("vars.INDEPENDENT_REVIEWER", workflow)
        self.assertIn('.state == "APPROVED"', workflow)
        self.assertIn(".commit_id ==", workflow)
        # The reviewer account is configured via repository variable, never
        # hard-coded in the workflow source.
        self.assertNotIn("policedbc", workflow)
        self.assertNotIn("cvsz", workflow)

    def test_no_obsolete_automated_approval_documentation(self):
        self.assertFalse(
            (ROOT / "docs/ai/playbooks/auto-audit-review-approve-policedbc.md")
            .exists(),
            "obsolete automated-approval documentation must not exist",
        )


if __name__ == "__main__":
    unittest.main()
