# kizami — CLAUDE.md

## Project Overview

A Go-based CLI tool (`kizami`) to maintain living documentation alongside code, with automatic drift detection.
Documents are saved as Markdown files under `docs/decisions/` (configurable) and managed with Git.

The core value: **the `## Related Files` section in any Markdown document links it to source files.
`kizami audit` detects when those source files are deleted or moved — keeping documentation honest.**

Supports any living document: ADRs, design docs, API specs, architecture overviews, and more.

---

## Directory Structure

```
kizami/
├── cmd/
│   ├── root.go         # Root command (kizami)
│   ├── log.go          # kizami adr / kizami design
│   ├── list.go         # kizami list
│   ├── search.go       # kizami search
│   ├── show.go         # kizami show
│   └── status.go       # kizami status
├── internal/
│   ├── decision/
│   │   ├── decision.go     # Decision type definition and parsing
│   │   ├── generate.go     # File generation and auto-numbering logic
│   │   └── decision_test.go
│   ├── search/
│   │   ├── search.go       # Keyword search
│   │   └── search_test.go
│   └── template/
│       └── template.go     # Markdown template management
├── docs/
│   └── decisions/          # ADRs for this repository itself (dogfooding)
│       ├── 2026-03-12-use-go-over-shell-script.md
│       ├── 2026-03-12-use-cobra-for-cli-framework.md
│       ├── 2026-03-12-madr-format-compatibility.md
│       ├── 2026-03-12-plaintext-markdown-only.md
│       └── 2026-03-12-ripgrep-fallback-strategy.md
├── CLAUDE.md
├── CLAUDE.ja.md
├── go.mod              # module github.com/mskasa/kizami
├── go.sum
└── main.go
```

---

## Tech Stack

| Purpose       | Library / Tool                          | Reason                                                          |
| ------------- | --------------------------------------- | --------------------------------------------------------------- |
| CLI framework | [cobra](https://github.com/spf13/cobra) | De facto standard for Go CLIs                                   |
| Testing       | Standard `go test`                      | Avoid unnecessary external dependencies                         |
| Search        | ripgrep (external command) + fallback   | Fast search; falls back to stdlib when ripgrep is not installed |
| Distribution  | GoReleaser + GitHub Actions             | Single binary distribution                                      |

- Go version: 1.22 or later
- Target OS: Linux / macOS / Windows (single binary)

---

## Command Specification (MVP)

```bash
kizami adr "<title>"              # Create a new ADR and open it in an editor
kizami design "<title>"           # Create a new design document and open it in an editor
kizami list                       # List decisions in reverse chronological order (slug, date, status, title)
kizami search <keyword>           # Search decisions by keyword
kizami show <slug>                # Display a single decision (e.g. kizami show use-go-over-shell-script)
kizami status <slug> <status>     # Update the status (e.g. kizami status use-postgresql superseded --by use-cockroachdb)
kizami blame <file>               # Find decisions related to a given file
```

### Status Definitions

| Status               | Meaning                              | When to use                                      |
| -------------------- | ------------------------------------ | ------------------------------------------------ |
| `Active`             | Currently valid decision (default)   | Commit together with the code change             |
| `Inactive`           | Simply no longer valid               | When no replacement ADR exists                   |
| `Superseded by <slug>` | Replaced by another decision       | When a new ADR is created to replace this one    |

**Status policy:**
- Default is `Active` — ADRs are committed alongside code changes, so the decision is considered final at creation time
- When a new ADR replaces an existing one, mark the old ADR as `Superseded by <slug>`
- When a decision becomes invalid without a replacement, mark it as `Inactive`

---

## Markdown Template (MADR-compatible)

Template generated when running `kizami adr`:

```markdown
# {Title}

- Date: {YYYY-MM-DD}
- Status: Draft
- Author: {git config user.name}

## Context

<!-- Why this decision was needed. Describe the background, constraints, and problem. -->

## Decision

<!-- What was decided. State clearly in 1–3 sentences. -->

## Consequences

<!-- Impact, benefits, and trade-offs of this decision. -->

## Alternatives Considered

<!-- Options that were considered but not adopted, and why. (Optional) -->

## Related Files

<!-- List files related to this decision (e.g. internal/search/search.go). -->
```

### File Naming Convention

```
YYYY-MM-DD-kebab-case-title.md
e.g. 2026-03-12-use-go-over-shell-script.md
```

- `YYYY-MM-DD`: creation date (preserves chronological sort order)
- kebab-case: title is automatically converted to lowercase with hyphens
- Saved under: `docs/decisions/` (relative to the repository root)
- For dogfooding ADRs in this repository, create both English and Japanese versions of each file:
  - English: `docs/decisions/2026-03-12-use-go-over-shell-script.md`
  - Japanese: `docs/decisions/ja/2026-03-12-use-go-over-shell-script.md`

---

## 🐕 Dogfooding Policy (Critical)

**This repository uses kizami itself to record its own design decisions.**

### Why Dogfooding Matters

- It is the strongest proof of value in the README ("the author actually uses this")
- Pain points discovered while writing ADRs become direct UX feedback for the tool
- Visitors to the GitHub repository can understand the tool's value just by browsing `docs/decisions/`

### Instructions for Claude

**During implementation, always propose creating an ADR when any of the following occurs:**

- Technology selection (libraries, algorithms, file formats)
- A choice between multiple implementation approaches
- Changing or retiring an existing design
- Any decision that affects future extensibility

**Example triggers:**

```
"Should we record the reason for choosing cobra in an ADR?"
"I'll create a Decision to document the ripgrep fallback strategy."
"This design choice is worth preserving in docs/decisions/."
```

### ADR Granularity Guidelines

**Decisions worth recording as an ADR:**

- Design decisions that affect multiple files or multiple components
- Decisions driven by external factors (load testing, incidents, performance measurements, etc.)
- Decisions that a future developer would want to understand ("why is it done this way?")

**Decisions that do NOT warrant an ADR:**

- Small-scale changes such as variable or function names
- Self-evident implementation details
- Reasons that are fully contained within a single file (write a code comment instead)

**ADR vs. Code Comment:**

| Scope | Where to document |
| ----- | ----------------- |
| Reasoning contained within a single file | Code comment |
| Reasoning that spans multiple files | ADR |
| Both | Write both; leave a link to the ADR in the comment |

Example — referencing an ADR from a code comment:

```go
// AuthorFromGit reads the author name from git config.
// Decision to use git config instead of an environment variable: docs/decisions/2026-03-16-allow-direct-adr-updates-with-git-history.md
func AuthorFromGit() string {
    ...
}
```

### ADR Update Policy

**ADRs can be updated directly, as Git manages the history.**
**Change history is tracked via `git log`.**

**Permitted changes:**
- Directly updating the content when the same decision is revised
  → `git diff` shows what changed; `git log` shows why
- Updating Status: `Active` → `Inactive` or `Superseded by <slug>`
- Fixing typos
- Appending entries to the Related Files section

**When to use Superseded:**
- When the direction of the decision changes entirely, create a new ADR and mark the old one as `Superseded by <slug>`
- When revising or refining the same decision, a direct update is sufficient

**Commit messages when updating an ADR:**
- Clearly state what was changed and why
- Good: `docs: update ADR madr-format-compatibility - increase pool size from 10 to 20 based on load test`
- Bad: `update adr`

### Initial ADRs to Create at Project Start

Before writing any code, manually create the following ADRs:

| Slug                       | Content                                                                           |
| -------------------------- | --------------------------------------------------------------------------------- |
| use-go-over-shell-script   | Why Go was chosen (single binary, Windows support, type safety)                   |
| use-cobra-for-cli-framework | Why cobra was chosen (de facto standard, shell completion, subcommand management) |
| madr-format-compatibility  | Why MADR format was adopted (compatibility with existing ADR tooling)             |
| plaintext-markdown-only    | Why plain Markdown was chosen over a database (Git-friendly, portable)            |
| ripgrep-fallback-strategy  | The decision around ripgrep dependency and fallback design                        |
| command-name-why           | Why the CLI command was originally named `why` (now superseded by rename-to-kizami-and-expand-scope) |

---

## Development Guidelines

### Coding Conventions

- Always run `gofmt` / `goimports` before committing
- Wrap errors using `fmt.Errorf("...: %w", err)`
- All CLI output messages must be in **English**
- Code comments must be written in **English**

### GitHub Actions

- Always pin actions to a full commit SHA instead of a version tag to prevent tag-mutation attacks
- Write the version tag first as a comment for readability, then run `pinact run` to convert it to a SHA:

```yaml
# Before
- uses: actions/checkout@v4

# After (run: pinact run)
- uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
```

#### Updating action versions

Run `pinact run --update` to update all actions in `.github/workflows/` to their latest versions:

```bash
pinact run --update
```

> **Note:** `pinact` only scans `.github/workflows/`. The workflow templates embedded in
> `internal/initializer/templates/` are **not** updated automatically — update them manually
> whenever the versions in `.github/workflows/` change.

### Testing Policy

- Place `_test.go` files in each package
- Use `t.TempDir()` for tests that involve file I/O
- Tests that depend on external commands (e.g. ripgrep) must include a skip condition:

```go
if _, err := exec.LookPath("rg"); err != nil {
    t.Skip("ripgrep not installed")
}
```

### Commit Message Convention

```
<type>: <summary>

Types:
  feat     New feature
  fix      Bug fix
  docs     Documentation (including ADR additions)
  refactor Refactoring
  test     Adding or updating tests
  chore    Build or dependency changes

Examples:
  feat: implement kizami adr command with auto-numbering
  docs: add ADR 0003 for MADR format compatibility
```

---

## Branch & PR Workflow

### Branch Strategy

Two branch types only — keep it simple for solo development:

```
main
└── feature/xxx   # one branch per feature, merged back to main when complete
```

No `develop` branch. It adds complexity without benefit for a solo project.

### Branch Naming

```bash
feature/kizami-log-command
feature/kizami-list-command
feature/auto-numbering
docs/initial-adrs           # ADR additions also get their own branch
fix/slug-generation-bug
```

### One Cycle: Start to Merge

Claude handles all Git operations. The owner approves and merges.

```
1. Claude creates a feature branch
2. Claude implements, commits, and pushes
3. Claude opens a PR (requires GitHub CLI: gh)
4. Owner reviews the PR and merges via GitHub UI or `gh pr merge`
5. Claude updates the Implementation Progress checklist in CLAUDE.md
```

Always use **squash merge** to keep main history clean.
Claude's working commits are implementation details — one clean commit per feature is enough.

### PR Description Template

When opening a PR, always include:

```markdown
## What
(One sentence summary of the change)

## Why
(Link to the relevant ADR if a design decision was involved)

## Checklist
- [ ] Tests pass (`go test ./...`)
- [ ] ADR created if a design decision was made
- [ ] CLAUDE.md progress updated
```

---

## How Claude and the Owner Work Together

Claude handles implementation. The owner handles judgment and approval.

| Task                                      | Who                        |
| ----------------------------------------- | -------------------------- |
| Create branch                             | Claude                     |
| Write code                                | Claude                     |
| Write tests                               | Claude                     |
| Commit and push                           | Claude                     |
| Open PR                                   | Claude (requires `gh` CLI) |
| Propose ADR when a design decision arises | Claude                     |
| **Review and approve PR**                 | **Owner**                  |
| **Merge to main**                         | **Owner**                  |
| **Confirm ADR content is accurate**       | **Owner**                  |
| **Decide what to work on next**           | **Owner**                  |

### Ideal Conversation Flow

```
Owner:
"Please read CLAUDE.md and understand the current state.
 Create branch feature/kizami-adr-command and implement the kizami adr command.
 Confirm with me at each step before proceeding."

Claude:
"Understood. I've read CLAUDE.md.
 Creating branch feature/kizami-adr-command now.
 [creates branch]
 Starting implementation of kizami adr...
 [implements]
 Done. The auto-numbering logic required a design choice.
 Shall I create an ADR before committing?"

Owner:
"Yes, please create the ADR."

Claude:
"Created docs/decisions/2026-03-23-auto-numbering-strategy.md.
 Committing and pushing now.
 Shall I open a PR?"

Owner:
"Yes."

Claude:
"PR opened: https://github.com/mskasa/kizami/pull/1
 Please review and merge when ready."
```

### Interaction Principles for Claude

- **Confirm before each major step** — do not proceed through branch → implement → commit → PR in one shot without pausing
- **Propose an ADR proactively** whenever a design decision is made during implementation
- **Keep tasks small** — one command per branch, one concern per PR
- **Update the Implementation Progress checklist** in CLAUDE.md after every merge

---

## Common Workflows

### Starting a New Feature

```
1. Owner says which feature to implement next
2. Claude creates the branch
3. Claude checks whether an ADR is needed before writing code
4. Claude implements and tests under internal/, then wires up cmd/
5. Claude commits, pushes, and opens a PR
6. Owner reviews and merges
7. Claude updates CLAUDE.md checklist
```

### Resuming Across Sessions

```
"Please read CLAUDE.md to understand the current state of the project.
 The last completed task was: [feature name].
 Next I'd like to implement: [next feature]."
```

### Scoping Requests

```
# Good — specific and bounded
"Implement only the auto-numbering logic in internal/decision/generate.go"
"Format the kizami list output using tabwriter"

# Avoid — too broad for a single session
"Implement the entire MVP"
```

---

## Roadmap

kizami is in active team use. The roadmap below reflects both OSS readiness goals and improvements driven by real-world feedback.

### Phase 1 — Legal & Trust Foundations ✅

*Prerequisites before promoting kizami to a wider audience.*

- [x] Add `LICENSE` file (MIT) — without it, use in many organizations is legally blocked
- [x] `SECURITY.md` — define how to report vulnerabilities privately
- [x] `CODE_OF_CONDUCT.md` — adopt Contributor Covenant or similar standard
- [x] `.github/ISSUE_TEMPLATE/` — bug report and feature request templates
- [x] `.github/PULL_REQUEST_TEMPLATE.md` — formalize the PR template currently in CLAUDE.md
- [x] Update `CONTRIBUTING.md` — golangci-lint version and setup instructions are out of date

### Phase 2 — Quality & Discoverability

*Raise the quality bar and make the project easier to find and trust.*

**Test coverage**
- [ ] Add regression tests as bugs are discovered from team usage

**CI**
- [x] Add Windows to the test matrix
- [ ] Add macOS to the test matrix (intentionally skipped for now; Linux + Windows is sufficient)

**GitHub repository**
- [x] Set repository Topics (`cli`, `golang`, `documentation`, `adr`, `decision-record`, `living-documentation`)
- [x] Add README badges (CI, Go Report Card, License, release)
- [ ] Expand README with team use-case stories and before/after examples

**Feature: noise reduction for long-running teams**
- [ ] `kizami archive` — move `Inactive` / `Superseded` documents to `docs/archive/` and exclude them from `kizami list`, `kizami audit`, and `kizami review`

### Phase 3 — Distribution & Ecosystem

*Broaden reach and make kizami adoptable by diverse teams.*

**Package managers**
- [ ] Homebrew formula
- [ ] Scoop (Windows)

**GitHub Actions**
- [ ] GitHub Actions Marketplace release (`kizami audit`, `kizami lint` as reusable actions)

**Extensibility**
- [ ] User-defined templates — configurable template path in `kizami.toml`

**Documentation site**
- [ ] Team onboarding guide
- [ ] Migration guide (from adr-tools, plain Markdown, Confluence/Notion)

---

### Team Feedback

kizami is in active team use. Feedback from real-world usage drives the roadmap.

- File issues with the `feedback` label on GitHub
- Pain points in `kizami lint` / `kizami audit` error messages are especially valuable
- Usability issues discovered during team use should be recorded as ADRs in `docs/decisions/` (dogfooding)

---

## References

- [MADR Format Specification](https://adr.github.io/madr/)
- [cobra Documentation](https://github.com/spf13/cobra)
- [adr-tools (reference implementation)](https://github.com/npryce/adr-tools)
- [GoReleaser](https://goreleaser.com/)
- [GitHub CLI (gh)](https://cli.github.com/) — required for Claude to open PRs
