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
│       ├── 0001-use-go-over-shell-script.md
│       ├── 0002-use-cobra-for-cli-framework.md
│       ├── 0003-madr-format-compatibility.md
│       ├── 0004-plaintext-markdown-only.md
│       └── 0005-ripgrep-fallback-strategy.md
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
kizami list                       # List decisions in reverse chronological order (ID, date, status, title)
kizami search <keyword>           # Search decisions by keyword
kizami show <id>                  # Display a single decision (e.g. kizami show 3)
kizami status <id> <status>       # Update the status (e.g. kizami status 3 superseded --by 5)
kizami blame <file>               # Find decisions related to a given file
```

### Status Definitions

| Status               | Meaning                              | When to use                                      |
| -------------------- | ------------------------------------ | ------------------------------------------------ |
| `Active`             | Currently valid decision (default)   | Commit together with the code change             |
| `Inactive`           | Simply no longer valid               | When no replacement ADR exists                   |
| `Superseded by NNNN` | Replaced by another decision         | When a new ADR is created to replace this one    |

**Status policy:**
- Default is `Active` — ADRs are committed alongside code changes, so the decision is considered final at creation time
- When a new ADR replaces an existing one, mark the old ADR as `Superseded by NNNN`
- When a decision becomes invalid without a replacement, mark it as `Inactive`

---

## Markdown Template (MADR-compatible)

Template generated when running `kizami adr`:

```markdown
# {NNNN}: {Title}

- Date: {YYYY-MM-DD}
- Status: Active
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
NNNN-kebab-case-title.md
e.g. 0001-use-go-over-shell-script.md
```

- `NNNN`: 4-digit zero-padded sequential number (auto-incremented from the current maximum)
- kebab-case: title is automatically converted to lowercase with hyphens
- Saved under: `docs/decisions/` (relative to the repository root)
- For dogfooding ADRs in this repository, create both English and Japanese versions of each file:
  - English: `docs/decisions/0001-use-go-over-shell-script.md`
  - Japanese: `docs/decisions/ja/0001-use-go-over-shell-script.md`

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
// Decision to use git config instead of an environment variable: docs/decisions/0009-author-source.md
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
- Updating Status: `Active` → `Inactive` or `Superseded by NNNN`
- Fixing typos
- Appending entries to the Related Files section

**When to use Superseded:**
- When the direction of the decision changes entirely, create a new ADR and mark the old one as `Superseded by NNNN`
- When revising or refining the same decision, a direct update is sufficient

**Commit messages when updating an ADR:**
- Clearly state what was changed and why
- Good: `docs: update ADR 0003 - increase pool size from 10 to 20 based on load test`
- Bad: `update adr`

### Initial ADRs to Create at Project Start

Before writing any code, manually create the following ADRs:

| ID   | Title                       | Content                                                                           |
| ---- | --------------------------- | --------------------------------------------------------------------------------- |
| 0001 | use-go-over-shell-script    | Why Go was chosen (single binary, Windows support, type safety)                   |
| 0002 | use-cobra-for-cli-framework | Why cobra was chosen (de facto standard, shell completion, subcommand management) |
| 0003 | madr-format-compatibility   | Why MADR format was adopted (compatibility with existing ADR tooling)             |
| 0004 | plaintext-markdown-only     | Why plain Markdown was chosen over a database (Git-friendly, portable)            |
| 0005 | ripgrep-fallback-strategy   | The decision around ripgrep dependency and fallback design                        |
| 0006 | command-name-why            | Why the CLI command was originally named `why` (now superseded by 0010)           |

---

## Development Guidelines

### Coding Conventions

- Always run `gofmt` / `goimports` before committing
- Wrap errors using `fmt.Errorf("...: %w", err)`
- All CLI output messages must be in **English**
- Code comments must be written in **English**

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
"Created docs/decisions/0007-auto-numbering-strategy.md.
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

## Implementation Progress

<!-- Update this checklist as work proceeds -->

### MVP (v0.1.0) ✅

- [x] .github/workflows/ci.yml (go test + go vet on every PR)
- [x] go.mod + cobra setup (`module github.com/mskasa/kizami`)
- [x] cmd/root.go (root `kizami` command)
- [x] internal/decision/generate.go (auto-numbering and file generation)
- [x] internal/template/template.go (Markdown template)
- [x] cmd/log.go (`kizami adr` / `kizami design`)
- [x] cmd/list.go (`kizami list`)
- [x] cmd/search.go (`kizami search`)
- [x] cmd/show.go (`kizami show`)
- [x] cmd/status.go (`kizami status`)
- [x] docs/decisions/ initial ADRs (0001–0006)
- [x] README.md
- [x] GoReleaser configuration

### v0.1.0 (remaining)

- [x] Logo image for README
- [x] cmd/blame.go (`kizami blame <file>` — full-text search for file path mentions in ADRs)
- [x] `kizami --version` — print version string

### v0.2.0

- [x] `kizami init` — initialize decisions directory
- [x] Auto-open editor after `kizami adr`
- [x] Suggest changed files (staged and unstaged) as Related Files candidates on `kizami adr`
- [x] Show similar ADR suggestions on `kizami adr` (keyword partial match)
- [x] `kizami list --status <status>` — filter list by status
- [x] `kizami supersede` — mark an ADR as superseded
- [x] `kizami review` — detect long-stale ADRs
- [x] Git hook to prompt ADR creation
- [x] GitHub Actions integration (`kizami init` generates workflow)

### v0.3.0

- [x] `kizami audit` — detect drift between Related Files and actual code
- [x] Scheduled CI run of `kizami audit` (weekly + auto GitHub Issue creation)
- [x] LLM-assisted ADR draft generation
- [x] `kizami init` generates `~/.config/kizami/config.toml` with default values

### Rename to kizami ✅

- [x] Rename GitHub repository: `mskasa/declog` → `mskasa/kizami`
- [x] Update `go.mod` module path: `github.com/mskasa/declog` → `github.com/mskasa/kizami`
- [x] Update all import paths across the codebase
- [x] Rename binary: `why` → `kizami` (cmd/root.go, .goreleaser.yaml)
- [x] Update config path: `~/.config/declog/` → `~/.config/kizami/`
- [x] Update README.md and README.ja.md
- [x] Update CLAUDE.md and CLAUDE.ja.md (reflect new identity)
- [x] Update existing ADRs that reference `why` command

### v0.4.0 (scope expansion)

- [x] `kizami adr` / `kizami design` — separate creation commands (replaces `kizami log --type`)
- [x] Design document template (saved under `docs/design/`, default `Status: Draft`)
- [x] Change ADR template default from `Status: Active` to `Status: Draft`
- [x] `kizami audit` skips `Draft` documents (only checks `Active`)
- [ ] `kizami init` generates optional auto-promote workflow (`kizami-promote.yml`): auto-promotes `Draft` → `Active` on push to main, with inline comments for customization
- [ ] `kizami audit` supports multiple directories (`audit.dirs` in config)
- [ ] Remove ADR-specific language from generic output messages

### v0.5.0 (audit enhancement)

- [ ] Drift detection beyond file existence (function/symbol level references)
- [ ] Generate reverse index (`.kizami/index.json`: file path → ADR IDs mapping) for faster `kizami blame` and external tool integration
- [ ] `kizami design --ai` — AI draft for design documents
- [ ] `kizami sync` — interactively update Related Files in existing documents

### v1.0.0 (public release)

- [ ] Documentation site (GitHub Pages)
- [ ] Homebrew formula
- [ ] golangci-lint in CI
- [ ] Tests for `cmd/` package
- [ ] Color output for `kizami list` and `kizami search`
- [ ] `kizami stats`
- [ ] GitHub Actions Marketplace release

---

## References

- [MADR Format Specification](https://adr.github.io/madr/)
- [cobra Documentation](https://github.com/spf13/cobra)
- [adr-tools (reference implementation)](https://github.com/npryce/adr-tools)
- [GoReleaser](https://goreleaser.com/)
- [GitHub CLI (gh)](https://cli.github.com/) — required for Claude to open PRs
