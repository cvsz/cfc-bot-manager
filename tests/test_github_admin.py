"""Tests for independent-review administration prerequisites."""

import unittest
from unittest.mock import patch

from scripts.github_admin import (
    REQUIRED_CHECKS,
    configured_independent_reviewer,
    protection_payload,
    validate_independent_reviewer,
)


class IndependentReviewerAdminTests(unittest.TestCase):
    def test_independent_approval_is_a_required_status_check(self):
        self.assertIn("verify-independent-approval", REQUIRED_CHECKS)
        payload = protection_payload()
        self.assertIn("verify-independent-approval", payload["required_status_checks"]["contexts"])
        self.assertTrue(payload["required_status_checks"]["strict"])

    def test_review_policy_keeps_review_and_codeowner_requirements(self):
        reviews = protection_payload()["required_pull_request_reviews"]
        self.assertGreaterEqual(reviews["required_approving_review_count"], 1)
        self.assertTrue(reviews["require_code_owner_reviews"])
        self.assertTrue(reviews["dismiss_stale_reviews"])
        self.assertTrue(reviews["require_last_push_approval"])

    def test_reviewer_must_be_a_valid_login_distinct_from_owner(self):
        self.assertEqual(validate_independent_reviewer("cvsz/ztemplate", "policedbc"), "policedbc")
        for reviewer in ("", "invalid/login", "cvsz", "bad name"):
            with self.subTest(reviewer=reviewer), self.assertRaises(ValueError):
                validate_independent_reviewer("cvsz/ztemplate", reviewer)

    @patch("scripts.github_admin.gh_api")
    def test_configured_reviewer_reads_repository_variable(self, gh_api):
        gh_api.return_value = {
            "variables": [
                {"name": "OTHER_VARIABLE", "value": "unused"},
                {"name": "INDEPENDENT_REVIEWER", "value": "policedbc"},
            ]
        }
        self.assertEqual(configured_independent_reviewer("cvsz/ztemplate"), "policedbc")
        gh_api.assert_called_once_with(
            "GET", "repos/cvsz/ztemplate/actions/variables?per_page=100"
        )

    @patch("scripts.github_admin.gh_api")
    def test_missing_reviewer_variable_is_not_assumed(self, gh_api):
        gh_api.return_value = {"variables": []}
        self.assertIsNone(configured_independent_reviewer("cvsz/ztemplate"))


if __name__ == "__main__":
    unittest.main()
