# Automatic updates from ztemplate

GitHub's **Use this template** creates an independent repository, not an ongoing upstream link. Generated repositories inherit the optional weekly sync workflow in `.github/workflows/template-sync.yml`.

## Operation

- Mondays 03:17 UTC, or manual `workflow_dispatch`: compare selected upstream governance prompts against the last successfully adopted template ref.
- Changes are proposed in a new pull request, **never automatically merged**.
- Only the three paths in `scripts/sync_template.py` are managed. Project-specific source, credentials, infrastructure, `AGENTS.md`, workflows, `CODEOWNERS`, and CI remain untouched.
- A locally edited managed file is not overwritten. The job fails closed and asks for manual reconciliation.
- The state file `.github/template-sync-state.json` records the upstream ref adopted on merge, not a deployed application version.
- An open sync PR blocks new sync proposals until reviewed and closed.
- All existing PR review/branch-protection controls continue to apply.

## Setup for generated repositories

1. Enable GitHub Actions and allow `GITHUB_TOKEN` to create pull requests where repository or organization policy permits. The workflow requests only job-scoped `contents: write` and `pull-requests: write`; repository-wide token defaults can remain read-only.
2. Keep `scripts/sync_template.py` and `.github/workflows/template-sync.yml` when creating your project.
3. Run `python3 scripts/sync_template.py --dry-run` locally for a preview.
4. Trigger **Propose upstream template updates** manually or use its weekly schedule.
5. Review conflicts, verify exact-head CI, and merge the PR only with an independent qualified reviewer and an authorized maintainer.

If organization policy denies token-created pull requests, the automation cannot create one; use a reviewed GitHub App token or manual PR rather than weakening repository governance. Runs initiated by `GITHUB_TOKEN` can have event-trigger limitations; validate downstream required checks on the PR head independently.

## Boundaries

This is a conservative **allowlisted sync**, not automatic propagation of every template file. Review broader upstream changes using the repository rollout guide before adopting additional files. Locally modified files always require manual reconciliation. The template repository itself skips the inherited scheduled job.
