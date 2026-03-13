# declog — CLAUDE.md

## Project Overview

A Go-based CLI tool to record and search architectural design decisions ("Why") with minimal friction.
Invoked as `why`, decisions are saved as Markdown files under `docs/decisions/` and managed with Git.

The goal is to make the reasoning behind design choices traceable — consolidating what tends to get scattered across Issues, PRs, and Slack into the repository itself.

---

## Directory Structure

```
declog/
├── cmd/
│   ├── root.go         # Root command (why)
│   ├── log.go          # why log
│   ├── list.go         # why list
│   ├── search.go       # why search
│   ├── show.go         # why show
│   └── status.go       # why status
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
├── go.mod              # module github.com/yourname/declog
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
why log "<title>"              # Generate a template Markdown file and open it in an editor
why list                       # List decisions in reverse chronological order (ID, date, status, title)
why search <keyword>           # Search decisions by keyword
why show <id>                  # Display a single decision (e.g. why show 3)
why status <id> <status>       # Update the status (e.g. why status 3 superseded --by 5)
why blame <file>               # Find decisions related to a given file (planned for future release)
```

### Status Definitions

| Status               | Meaning                              | When to use                                      |
| -------------------- | ------------------------------------ | ------------------------------------------------ |
| `Active`             | Currently valid decision (default)   | Commit together with the code change             |
| `Inactive`           | Simply no longer valid               | When no replacement ADR exists                   |
| `Superseded by NNNN` | Replaced by another decision         | When a new ADR is created to replace this one    |

**Status policy:**
- Default is `Active` — since ADRs are committed alongside code changes, the decision is considered final at creation time
- When a new ADR replaces an existing one, mark the old ADR as `Superseded by NNNN`
- When a decision becomes invalid without a replacement, mark it as `Inactive`

---

## Markdown Template (MADR-compatible)

Template generated when running `why log`:

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

**This repository uses declog itself to record its own design decisions.**

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

### Initial ADRs to Create at Project Start

Before writing any code, manually create the following ADRs:

| ID   | Title                       | Content                                                                           |
| ---- | --------------------------- | --------------------------------------------------------------------------------- |
| 0001 | use-go-over-shell-script    | Why Go was chosen (single binary, Windows support, type safety)                   |
| 0002 | use-cobra-for-cli-framework | Why cobra was chosen (de facto standard, shell completion, subcommand management) |
| 0003 | madr-format-compatibility   | Why MADR format was adopted (compatibility with existing ADR tooling)             |
| 0004 | plaintext-markdown-only     | Why plain Markdown was chosen over a database (Git-friendly, portable)            |
| 0005 | ripgrep-fallback-strategy   | The decision around ripgrep dependency and fallback design                        |
| 0006 | command-name-why            | Why the CLI command was named `why` instead of `dec` or `declog`                  |

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
  feat: implement why log command with auto-numbering
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
feature/why-log-command
feature/why-list-command
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
 Create branch feature/why-log-command and implement the why log command.
 Confirm with me at each step before proceeding."

Claude:
"Understood. I've read CLAUDE.md.
 Creating branch feature/why-log-command now.
 [creates branch]
 Starting implementation of why log...
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
"PR opened: https://github.com/yourname/declog/pull/1
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
"Format the why list output using tabwriter"

# Avoid — too broad for a single session
"Implement the entire MVP"
```

---

## Implementation Progress

<!-- Update this checklist as work proceeds -->

### MVP (v0.1.0) ✅

- [x] .github/workflows/ci.yml (go test + go vet on every PR)
- [x] go.mod + cobra setup (`module github.com/mskasa/declog`)
- [x] cmd/root.go (root `why` command)
- [x] internal/decision/generate.go (auto-numbering and file generation)
- [x] internal/template/template.go (Markdown template)
- [x] cmd/log.go (`why log`)
- [x] cmd/list.go (`why list`)
- [x] cmd/search.go (`why search`)
- [x] cmd/show.go (`why show`)
- [x] cmd/status.go (`why status`)
- [x] docs/decisions/ initial ADRs (0001–0006)
- [x] README.md
- [x] GoReleaser configuration

### Near-term (v0.1.x)

- [ ] Demo GIF in README
- [ ] cmd/blame.go (`why blame <file>` — full-text search for file path mentions in ADRs)
- [ ] cmd/edit.go (`why edit <id>` — open an existing ADR in `$EDITOR`)
- [ ] `why list --status <status>` — filter list by status
- [ ] `why search -i` — case-insensitive search flag
- [ ] `why --version` — print version string
- [ ] Tests for `cmd/` package

### v0.2.0

- [ ] Homebrew formula
- [ ] Scoop manifest (Windows)
- [ ] Custom decisions directory (`--dir` flag or config file)

### Post-MVP improvements

- [ ] Color output for `why list` and `why search`
- [ ] golangci-lint in CI

---

## References

- [MADR Format Specification](https://adr.github.io/madr/)
- [cobra Documentation](https://github.com/spf13/cobra)
- [adr-tools (reference implementation)](https://github.com/npryce/adr-tools)
- [GoReleaser](https://goreleaser.com/)
- [GitHub CLI (gh)](https://cli.github.com/) — required for Claude to open PRs
