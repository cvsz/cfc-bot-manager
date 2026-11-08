"""Static contract tests for the independent-review workflow and rollout script."""

import os
from pathlib import Path
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[1]
WORKFLOW = (ROOT / ".github/workflows/request-independent-review.yml").read_text(encoding="utf-8")
ROLLOUT = (ROOT / "scripts/rollout-independent-review.sh").read_text(encoding="utf-8")


class IndependentReviewWorkflowTests(unittest.TestCase):
    def test_review_request_is_refreshed_on_head_updates(self):
        self.assertIn("synchronize", WORKFLOW)
        self.assertIn("pull_request_review:", WORKFLOW)
        self.assertIn("[submitted, dismissed]", WORKFLOW)

    def test_gate_requires_a_distinct_reviewer_on_the_exact_head(self):
        self.assertIn("verify-independent-approval:", WORKFLOW)
        self.assertIn("vars.INDEPENDENT_REVIEWER", WORKFLOW)
        self.assertIn("HEAD_SHA:", WORKFLOW)
        self.assertIn('.state == "APPROVED"', WORKFLOW)
        self.assertIn(".commit_id == $sha", WORKFLOW)
        self.assertIn("PR_AUTHOR", WORKFLOW)
        self.assertNotIn("policedbc", WORKFLOW)
        self.assertNotIn("actions/checkout", WORKFLOW)

    def test_permissions_are_scoped_and_rollout_skips_mirrors(self):
        self.assertIn("pull-requests: read", WORKFLOW)
        self.assertIn("pull-requests: write", WORKFLOW)
        self.assertIn("mirror_url", ROLLOUT)
        self.assertIn("OWNER=", ROLLOUT)
        self.assertNotIn(chr(92) + "$" + "{", ROLLOUT)
        self.assertNotIn(chr(92) + "$" + "{{", WORKFLOW)
        self.assertIn('cp "$SOURCE_WORKFLOW" "$WORKFLOW"', ROLLOUT)
        self.assertIn("INDEPENDENT_REVIEWER", ROLLOUT)
        self.assertNotIn("REVIEWER:-policedbc", ROLLOUT)

    def test_rollout_uses_authenticated_clone_for_private_repositories(self):
        self.assertIn('gh repo clone "$repo" "$repo_dir" -- --depth 1 --branch "$base"', ROLLOUT)
        self.assertNotIn('git clone --quiet --depth 1 --branch "$base" "https://github.com/$repo.git"', ROLLOUT)

    def test_rollout_executes_authenticated_clone_for_private_repository(self):
        with tempfile.TemporaryDirectory() as tmp:
            temp = Path(tmp)
            bin_dir = temp / "bin"
            bin_dir.mkdir()
            call_log = temp / "calls.log"
            gh = bin_dir / "gh"
            gh.write_text(
                "#!/usr/bin/env bash\n"
                "set -euo pipefail\n"
                "printf '%s\\n' \"$*\" >> \"$CALL_LOG\"\n"
                "if [[ \"$1\" == api && \"$2\" == users/cvsz ]]; then printf 'User\\n'; exit 0; fi\n"
                "if [[ \"$1\" == api && \"$2\" == --paginate ]]; then printf 'cvsz/private-repo\\tmain\\tfalse\\tfalse\\tfalse\\t\\n'; exit 0; fi\n"
                "if [[ \"$1\" == api && \"$2\" == repos/cvsz/private-repo/collaborators/policedbc/permission ]]; then printf 'admin\\n'; exit 0; fi\n"
                "if [[ \"$1\" == api ]]; then case \"$2\" in repos/cvsz/private-repo/contents/*|repos/cvsz/private-repo/branches/*) exit 1 ;; esac; fi\n"
                "if [[ \"$1\" == repo && \"$2\" == clone ]]; then mkdir -p \"$4\"; exit 0; fi\n"
                "if [[ \"$1\" == pr && \"$2\" == create ]]; then exit 0; fi\n"
                "exit 0\n",
                encoding="utf-8",
            )
            gh.chmod(0o755)
            git = bin_dir / "git"
            git.write_text(
                "#!/usr/bin/env bash\n"
                "set -euo pipefail\n"
                "printf 'git %s\\n' \"$*\" >> \"$CALL_LOG\"\n",
                encoding="utf-8",
            )
            git.chmod(0o755)

            env = os.environ.copy()
            env.update(
                {
                    "PATH": f"{bin_dir}{os.pathsep}{env.get('PATH', '')}",
                    "OWNER": "cvsz",
                    "REVIEWER": "policedbc",
                    "APPLY": "1",
                    "CALL_LOG": str(call_log),
                }
            )
            result = subprocess.run(
                ["bash", str(ROOT / "scripts/rollout-independent-review.sh")],
                capture_output=True,
                text=True,
                env=env,
                check=False,
            )
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            calls = call_log.read_text(encoding="utf-8")
            self.assertIn("repo clone cvsz/private-repo", calls)
            self.assertIn("--depth 1 --branch main", calls)
            self.assertNotIn("git clone", calls)
            self.assertIn("PR CREATED cvsz/private-repo", result.stdout)

    def test_rollout_selects_user_or_organization_repository_endpoint(self):
        self.assertIn("owner_type=", ROLLOUT)
        self.assertIn("Organization)", ROLLOUT)
        self.assertIn("orgs/$OWNER/repos?per_page=100&type=all", ROLLOUT)
        self.assertIn("users/$OWNER/repos?per_page=100&type=owner", ROLLOUT)


if __name__ == "__main__":
    unittest.main()
