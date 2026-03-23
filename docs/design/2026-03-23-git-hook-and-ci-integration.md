# Git Hook and CI Integration

- Date: 2026-03-23
- Type: Design
- Status: Active
- Author: masahiro.kasatani

## Overview

kizami integrates with the development workflow at two points: a local pre-commit hook that warns when a commit lacks a document, and a GitHub Actions workflow (`adr-check.yml`) that checks each pull request for document coverage. Both are opt-in and installed by `kizami init`.

## Background

The primary risk for any documentation tool is that developers forget to write documents alongside code changes. Automated reminders at commit and PR time close this gap without requiring a manual process or a review checklist item. Both touchpoints are designed as soft gates — they warn rather than block — to avoid friction while still surfacing the reminder at the moment it is most relevant.

## Goals / Non-Goals

**Goals:**
- Remind developers to create an ADR or design document at commit time (pre-commit hook)
- Check at PR time whether the PR includes a document or references one (CI workflow)
- Provide an escape hatch (`[skip-doc]` in PR title) for PRs that genuinely don't need a document
- Be opt-in: installed only when the user confirms during `kizami init`
- Handle pre-existing hooks gracefully (print content for manual append rather than overwriting)

**Non-Goals:**
- Hard-blocking commits or PRs — both integrations warn only
- Tracking which ADRs are "linked" to which PRs in a structured way
- Requiring a specific ADR format in the PR description

## Design

### Pre-Commit Hook

The hook script is a POSIX-compatible shell script (`templates/pre-commit`), embedded in the binary via `//go:embed` and written to `.git/hooks/pre-commit` by `InstallHook` in `internal/initializer/hook.go`.

#### Skip Conditions

The hook exits 0 (no warning) if any of the following is true:

1. **Document files are staged**: `git diff --cached --name-only` contains a path under `docs/decisions/`
2. **Documentation-only commit**: all staged files have a `.md` extension — this is a docs update, not a code change requiring a new document

If neither condition is met and non-`.md` files are staged, the warning is printed:

```
⚠️  No ADR found in this commit.
    If this change involves a significant design decision,
    consider running: kizami adr "<title>"
```

The hook does **not** fail (exit 1) — it always exits 0. This is intentional: a hard block would cause friction for small bug fixes or typo corrections that don't warrant a document.

#### Pre-existing Hook Handling

If `.git/hooks/pre-commit` already exists when `kizami init` runs, `InstallHook` does not overwrite it. Instead, it prints the hook script content to stdout with a message asking the user to append it manually. Silently overwriting an existing hook would break other tooling (e.g., linters, formatters) without warning.

### CI Workflow (adr-check.yml)

The GitHub Actions workflow runs on every pull request (`opened`, `edited`, `synchronize` events).

#### Check Logic

The workflow passes (exits 0) if **any** of the following is true:

1. **`[skip-doc]` in PR title**: the author explicitly signals that no document is needed
2. **Document path in PR body**: the PR description references `docs/decisions/` or `docs/design/` — indicating the PR is linked to an existing document
3. **Document file in changed files**: `git diff --name-only BASE..HEAD` contains a path matching `^docs/(decisions|design)/.*\.md$`

If none of the above is true, the workflow emits a GitHub Actions warning annotation:

```
::warning::No document found. Consider adding a decision record to docs/decisions/ or a design doc to docs/design/, or include [skip-doc] in the PR title.
```

The warning annotation appears in the PR checks UI but does not set the check to "failed". This keeps it non-blocking while still being visible.

#### Why a Warning, Not a Failure

Making document coverage a hard requirement would:
- Block hotfixes and urgent PRs
- Generate noise for refactors and dependency bumps that don't involve design decisions
- Frustrate teams adopting kizami incrementally

A warning is visible enough to prompt action for significant changes, while not creating obstacles for routine work.

### Relationship Between Hook and CI

The two integrations are complementary and independent:

| | Pre-commit hook | CI workflow |
|---|---|---|
| Trigger | `git commit` | Pull request |
| Scope | Staged files | All commits in PR |
| Check | `docs/decisions/` staged | `docs/(decisions\|design)/` in diff or PR body |
| Escape hatch | None (warning always shown if conditions aren't met) | `[skip-doc]` in PR title |
| Blocking | Never | Never |

The hook catches the issue as early as possible (at commit time); the CI workflow catches it at the integration point (when the PR is opened or updated).

## Open Questions

- **Configurable docs path**: The hook and workflow hardcode `docs/decisions/` and `docs/design/`. If a team uses a different path, they must edit the generated files manually. A future version could template the paths from `kizami.toml` during `kizami init`.
- **Hard-block mode**: Some teams may want a strict policy that fails PRs without documents. An opt-in `strict: true` flag in the workflow could change the warning to an error.
- **Hook coverage for `docs/design/`**: The pre-commit hook only checks for `docs/decisions/` files, not `docs/design/`. This is a known inconsistency with the CI workflow.

## Related Files

- `internal/initializer/hook.go`
- `internal/initializer/templates/pre-commit`
- `internal/initializer/templates/adr-check.yml`
- `internal/initializer/init.go`
- `cmd/init.go`
