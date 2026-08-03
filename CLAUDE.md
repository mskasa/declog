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

- Go version: 1.25 or later
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
- Run `golangci-lint run` before pushing and fix all reported issues
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
- [x] Expand README with team use-case stories and before/after examples

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

### Phase 4 — Agent Context Layer ✅

*Move kizami's center of gravity from "a human actively runs the CLI" to "the right decision surfaces automatically at the right moment." AI agents can already read `docs/decisions/` directly, so reachability was never the gap — the gap is that decisions aren't reliably written at the moment they're made, and aren't reliably surfaced at the moment they matter.*

**Step 1 — Context resolver** ✅
- [x] `internal/context` package: unify the two existing "related files" implementations (`search.Blame`'s full-text search and `decision.CheckHook`'s structured `## Related Files` parsing) into a single definition
- [x] Add glob support (e.g. `internal/**/*_test.go`) to Related Files entries, alongside the existing exact-path and directory-prefix matching
- [x] `kizami context <files...> [--json] [--full]` — given a set of changed files, return the Active decisions that govern them (plus Superseded decisions with their `supersededBy` target), and per-file drift state

**Step 2 — Agent manifest sync** ✅
- [x] `kizami agents sync` — maintain a marker-delimited section in CLAUDE.md / AGENTS.md listing which paths are governed by which decision (a pointer table, not full ADR text)
- [x] `kizami agents sync --check` — CI check that fails when an ADR's Related Files aren't reflected in the manifest

**Step 3 — MCP server** ✅
- [x] `kizami mcp` — expose the resolver as MCP tools framed as questions an agent asks, not 1:1 mirrors of CLI verbs: `kizami_decisions_for_files`, `kizami_search_decisions`, `kizami_get_decision`
- [x] Default tool responses to the `## Decision` summary only (`full` param to escalate to full text) to keep agent context budget under control

**Step 4 — Agent-authored decisions** ✅
- [x] `kizami_record_decision` MCP tool (write path) — lets an agent record a decision immediately after making it, always as `Status: Draft`, new-file-only (no edits/deletes of existing docs or code), gated behind `kizami mcp --allow-write`
- [x] `kizami hook pre-tool-use` — Claude Code tool hook that injects `kizami context` output before an Edit/Write on a governed file, as a deterministic fallback when an agent doesn't proactively read the manifest or call the MCP tools

*Expected to span multiple ADRs (see dogfooding policy) — at minimum: the resolver/injection strategy, the "questions not verbs" MCP tool design, and the Related Files definition unification.*

### Backlog — Candidate Features

*Ideas that emerged from real usage but are not yet scheduled into a phase.*

- [ ] `kizami list --status <status>` — filter list output by status (e.g. `--status active`) to reduce noise as documents accumulate over time

### Deferred Simplification (from the pre-main-merge audit of Phase 4)

*Identified during a "simple is best" audit before merging `feature/agent-context-layer` into `main`. Deferred by the owner, not forgotten — action items, not just observations.*

- [ ] Remove `decision.FindBySlug`; have its two remaining callers (`cmd/log.go`'s `--supersedes` resolution) use `FindAllBySlug` and take the first result. The two functions are near-duplicate ~35-line `filepath.WalkDir` implementations differing only in early-exit behavior.
- [ ] Remove the dead `[warn]` severity branch in `kizami lint` (`cmd/lint.go`'s `warnCount` tracking and printing, `internal/decision/lint.go`'s `Severity` field/comment) — left over from the `lint-empty-related-files-as-warning` → `remove-empty-related-files-lint-warning` flip-flop; nothing sets `Severity: "warning"` anymore.
- [ ] Remove `--ai` draft generation entirely (`cmd/log.go`'s `--ai`/`--model`/`--dry-run` flags, `internal/ai/` package — both Anthropic and Bedrock backends, `[ai]` config section). Rationale: its whole job (produce a decision draft informed by code context) is now done strictly better by an AI coding agent working in the repo directly — richer context (full session, not a 2000-char diff truncation), zero extra API call, zero extra backend to maintain. Predates the Agent Context Layer (Phase 4) and was designed for a world without an agent already in the loop.
- [ ] Remove `kizami_record_decision` and the `--allow-write` flag/gate from `kizami mcp` (`internal/context/record.go`, the tool registration in `cmd/mcp.go`, and `docs/decisions/2026-08-03-agent-authored-decision-write-path.md`'s implementation — keep the ADR itself as a record of why it was tried and retired). Rationale: its safety model (new-file-only, Draft-only) isn't actually enforced — an agent with normal file-write access (which it needs anyway) can bypass it entirely by writing directly, which is exactly what happened throughout the session that built it. Everything it provides is already achieved by a CLAUDE.md/AGENTS.md convention ("write new decisions as `Status: Draft`, following the template") plus the agent's native Write tool, at zero additional operational cost (no `kizami mcp --allow-write` to run). Keep the 3 read-only MCP tools — they survive the same scrutiny because the underlying matching/resolution logic is genuinely error-prone to hand-roll (five real bugs were found while building it) and is cheaper in tokens than an agent re-deriving it from raw file reads each time.

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

<!-- kizami:start -->
<!-- Generated by `kizami agents sync` — do not edit by hand. -->
## Decisions Governing This Codebase

This repository's design decisions are recorded in `docs/decisions/` (and `docs/design/`). Read the decision below before editing files it covers.

| Related Files | Decision | Slug |
|---|---|---|
| `cmd/mcp.go`, `internal/context/record.go` | **既定でオフ、明示的なオプトイン。** `kizami mcp` は `--allow-write` を付けて起動した場合のみ `kizami_record_decision` を登録する。フラグがなければ、そのツールはエージェントの視点からは存在しない——登録はされているが拒否するのではなく、本当に存在しない。呼び… | agent-authored-decision-write-path |
| `cmd/mcp.go`, `internal/context/record.go` | **Off by default, explicit opt-in.** `kizami mcp` only registers `kizami_record_decision` when started with `--allow-write`. Without the flag, the tool doesn't… | agent-authored-decision-write-path |
| `internal/context/`, `cmd/context.go`, `cmd/agents.go`, `internal/decision/match.go`, `internal/decision/decision.go`, `internal/decision/audit.go`, `internal/decision/hook.go`, `internal/search/blame.go`, `internal/config/config.go` | kizami currently delivers value only when a human actively runs the CLI (`show`, `blame`, `search`). This design moves the center of gravity to "the right decis… | agent-context-layer-design |
| `internal/context/`, `cmd/context.go`, `cmd/agents.go`, `internal/decision/match.go`, `internal/decision/decision.go`, `internal/decision/audit.go`, `internal/decision/hook.go`, `internal/search/blame.go`, `internal/config/config.go` | kizami は現状、人間が能動的にCLI（`show`・`blame`・`search`）を打った時にのみ価値を発揮する。本設計は重心を「必要な瞬間に決定が自動で現れる」へ移す——特にAIコーディングエージェント向けに。kizami の `## Related Files` によるリンク付けとドリフト検知が最も効くの… | agent-context-layer-design |
| `internal/context/manifest.go`, `internal/context/sync.go`, `cmd/agents.go`, `internal/config/config.go` | **Delimited, idempotent block.** The generated content is wrapped in HTML comment markers: ``` <!-- kizami:start --> ... <!-- kizami:end --> ``` `kizami agents… | agent-manifest-sync-format |
| `internal/context/manifest.go`, `internal/context/sync.go`, `cmd/agents.go`, `internal/config/config.go` | **区切られた、べき等なブロック。** 生成されるコンテンツはHTMLコメントのマーカーで囲む。 ``` <!-- kizami:start --> ... <!-- kizami:end --> ``` `kizami agents sync` は、既存のマーカー対が見つかればその間のコンテンツをそのまま置き換える。… | agent-manifest-sync-format |
| `CLAUDE.md` | ADRはGitの履歴管理を前提に直接更新を許容する。何が変わったか・なぜ変えたかの根拠はGit履歴が担う。 運用ガイドライン： - 同じ判断を修正・精緻化する場合はADRをそのまま更新する - `git diff` で変更内容を伝え、コミットメッセージで変更理由を説明する - ADR更新時のコミットメッセージには何をな… | allow-direct-adr-updates-with-git-history |
| `CLAUDE.md` | ADRs can be updated directly. Git history is the source of truth for what changed and why. Guidelines: - When refining or correcting the same decision, update t… | allow-direct-adr-updates-with-git-history |
| `internal/decision/audit.go`, `internal/search/blame.go`, `cmd/audit.go`, `internal/initializer/templates/kizami-audit.yml`, `kizami.toml` | `kizami audit` detects when source files listed in a document's `## Related Files` section no longer exist in the repository. This keeps living documents honest… | audit-and-drift-detection |
| `internal/decision/audit.go`, `internal/search/blame.go`, `cmd/audit.go`, `internal/initializer/templates/kizami-audit.yml`, `kizami.toml` | `kizami audit` は、ドキュメントの `## Related Files` セクションに記載されたソースファイルがリポジトリ上に存在しなくなったことを検出します。 これにより、開発者が手動で更新することを覚えていなくても、陳腐化した参照を自動的に発見し、ドキュメントの正確性を維持します。 | audit-and-drift-detection |
| `internal/ai/bedrock.go`, `internal/ai/draft.go`, `internal/config/config.go`, `cmd/log.go` | AWS Bedrockを代替AIバックエンドとして追加する。バックエンドの選択は以下の優先順位に従う： 1. 環境変数 `CLAUDE_CODE_USE_BEDROCK=1` → Bedrock（Claude Code互換） 2. `kizami.toml` の `[ai] backend = "bedrock"` →… | bedrock-backend-support |
| `internal/ai/bedrock.go`, `internal/ai/draft.go`, `internal/config/config.go`, `cmd/log.go` | Add AWS Bedrock as an alternative AI backend. Backend selection follows this priority: 1. `CLAUDE_CODE_USE_BEDROCK=1` environment variable → Bedrock (Claude Cod… | bedrock-backend-support |
| `main.go`, `.goreleaser.yaml` | Name the CLI command `why` instead of `dec`, `declog`, or `adr`. "why" directly expresses the intent of the tool: recording the reasoning behind decisions. | command-name-why |
| `main.go`, `.goreleaser.yaml` | CLIコマンド名を`dec`・`declog`・`adr`ではなく`why`にする。 「why」はこのツールの目的——意思決定の理由を記録すること——を直接表現している。 | command-name-why |
| `internal/decision/generate.go`, `internal/decision/decision_test.go` | Recognise any `.md` file as a kizami document if it contains **both** of the following markers: 1. A line beginning with `- Status:` 2. A `## Related Files` sec… | content-based-document-detection |
| `internal/decision/generate.go`, `internal/decision/decision_test.go` | 以下の **両方** のマーカーを含む `.md` ファイルを kizami ドキュメントとして認識する： 1. `- Status:` で始まる行 2. `## Related Files` セクション見出し（ドリフト検出に必須） 両方のマーカーが存在する場合のみ対象となる。どちらか一方しか含まないファイルは kiz… | content-based-document-detection |
| `internal/decision/generate.go`, `cmd/root.go` | `decision.FindAllBySlug(dir, slug) ([]*Decision, error)` を追加する：`FindBySlug` と同じ再帰的な走査だが、早期終了なしに全マッチを収集する。`FindBySlug` 自体は変更せず、既存の挙動と呼び出し元をそのまま維持する。 `cmd.findAll… | findbyslug-recursive-language-variants |
| `internal/decision/generate.go`, `cmd/root.go` | Add `decision.FindAllBySlug(dir, slug) ([]*Decision, error)`: the same recursive walk as `FindBySlug`, but without the early exit — it collects every match inst… | findbyslug-recursive-language-variants |
| `internal/initializer/hook.go`, `internal/initializer/templates/pre-commit`, `internal/initializer/templates/kizami-check.yml`, `internal/initializer/init.go`, `cmd/init.go` | kizami integrates with the development workflow at two points: a local pre-commit hook that warns when a commit lacks a document, and a GitHub Actions workflow… | git-hook-and-ci-integration |
| `internal/initializer/hook.go`, `internal/initializer/templates/pre-commit`, `internal/initializer/templates/kizami-check.yml`, `internal/initializer/init.go`, `cmd/init.go` | kizami は開発ワークフローの2つのポイントで統合される。ドキュメントなしのコミットを警告するローカルの pre-commit フックと、プルリクエストのドキュメントカバレッジをチェックする GitHub Actions ワークフロー（`kizami-check.yml`）。どちらもオプトインで `kizami i… | git-hook-and-ci-integration |
| `internal/initializer/templates/pre-commit`, `internal/initializer/hook.go`, `internal/decision/hook.go`, `cmd/hook.go` | Move all pre-commit hook logic into a Go command: `kizami hook pre-commit`. The shell script template becomes a thin, portable wrapper: ```sh #!/bin/sh if comma… | hook-command-go-binary |
| `internal/initializer/templates/pre-commit`, `internal/initializer/hook.go`, `internal/decision/hook.go`, `cmd/hook.go` | pre-commitフックのロジック全体をGoのサブコマンド `kizami hook pre-commit` に移行する。 シェルスクリプトのテンプレートは、薄いラッパーに変更する： ```sh #!/bin/sh if command -v kizami >/dev/null 2>&1; then kizami h… | hook-command-go-binary |
| `internal/initializer/init.go`, `internal/initializer/hook.go`, `internal/initializer/templates/kizami-check.yml`, `internal/initializer/templates/kizami-audit.yml`, `internal/initializer/templates/kizami-promote.yml`, `internal/initializer/templates/pre-commit`, `cmd/init.go` | `kizami init` は、1つのインタラクティブなコマンドでリポジトリに kizami を導入するためのセットアップを行う。decisions ディレクトリの作成、`kizami.toml` の生成、オプションで CI ワークフローと pre-commit フックのインストールまでをカバーする。 | kizami-init |
| `internal/initializer/init.go`, `internal/initializer/hook.go`, `internal/initializer/templates/kizami-check.yml`, `internal/initializer/templates/kizami-audit.yml`, `internal/initializer/templates/kizami-promote.yml`, `internal/initializer/templates/pre-commit`, `cmd/init.go` | `kizami init` bootstraps a repository for kizami usage in a single interactive command: it creates the decisions directory, generates `kizami.toml`, and optiona… | kizami-init |
| `internal/decision/review.go`, `cmd/review.go`, `internal/config/config.go` | `kizami review` は、git のコミット履歴を「最終更新日時」の情報源として、設定可能な期間以上更新されていない Active なドキュメントを一覧表示する。 | kizami-review |
| `internal/decision/review.go`, `cmd/review.go`, `internal/config/config.go` | `kizami review` surfaces Active documents that have not been updated in a configurable number of months, using git commit history as the source of truth for "la… | kizami-review |
| `internal/decision/lint.go`, `internal/decision/lint_test.go`, `cmd/lint.go` | `kizami lint` は Markdown ドキュメントの `## Related Files` セクションが存在しない・空の場合、 `[error]` ではなく `[warn]` を出力するよう変更した。 警告のみの場合は exit code 0 で終了するため、CI はブロックされない。 sidecar（`.… | lint-empty-related-files-as-warning |
| `internal/decision/lint.go`, `internal/decision/lint_test.go`, `cmd/lint.go` | `kizami lint` now emits a `[warn]` line instead of `[error]` when a Markdown document's `## Related Files` section is missing or empty. The process exits with c… | lint-empty-related-files-as-warning |
| `internal/ai/draft.go`, `internal/ai/prompt.go`, `internal/decision/generate.go`, `internal/template/template.go`, `cmd/log.go`, `internal/config/config.go` | `kizami adr --ai` および `kizami design --ai` は、Anthropic API を使って現在の git コンテキスト（変更ファイルとステージ済み diff）からドキュメントのドラフトを生成する。コード変更と同時にリビングドキュメントを書く際の摩擦を減らすことが目的。 | llm-assisted-draft-generation |
| `internal/ai/draft.go`, `internal/ai/prompt.go`, `internal/decision/generate.go`, `internal/template/template.go`, `cmd/log.go`, `internal/config/config.go` | `kizami adr --ai` and `kizami design --ai` use the Anthropic API to generate draft document sections from the current git context (changed files and staged diff… | llm-assisted-draft-generation |
| `cmd/mcp.go`, `internal/context/search.go`, `internal/context/get.go` | Three read-only tools for this step (a fourth, write-capable tool is deferred to Step 4 — see [[agent-context-layer-design]]): **`kizami_decisions_for_files`**… | mcp-tools-as-questions-not-verbs |
| `cmd/mcp.go`, `internal/context/search.go`, `internal/context/get.go` | 本ステップでは読み取り専用のツールを3つとする（書き込み可能な4本目のツールはStep 4に持ち越す——[[agent-context-layer-design]] 参照）。 **`kizami_decisions_for_files`** — 「これから読む・編集するファイルを縛っているものは何か」。`interna… | mcp-tools-as-questions-not-verbs |
| `cmd/hook.go`, `internal/decision/hook.go` | `kizami hook pre-tool-use` は標準入力から `PreToolUse` のJSONを読み、`tool_input.file_path` を抽出してリポジトリルート相対パスに解決し、本フェーズの他のすべてのツールが既に使っている同じ `internal/context.Resolve` を呼ぶ。何… | pre-tool-use-hook-context-injection |
| `cmd/hook.go`, `internal/decision/hook.go` | `kizami hook pre-tool-use` reads the `PreToolUse` JSON from stdin, extracts `tool_input.file_path`, resolves it to a repo-relative path, and calls the same `int… | pre-tool-use-hook-context-injection |
| `internal/decision/match.go`, `internal/decision/audit.go`, `internal/decision/hook.go`, `internal/search/blame.go`, `internal/context/` | `internal/decision`（既に `ParseRelatedFiles` を持つパッケージ）に小さな共有プリミティブを2つ追加し、既存の全呼び出し箇所をこちらに寄せる。 **`Match(entry, file string) (kind MatchKind, ok bool)`** — 「Related… | related-files-single-definition |
| `internal/decision/match.go`, `internal/decision/audit.go`, `internal/decision/hook.go`, `internal/search/blame.go`, `internal/context/` | Add two small shared primitives to `internal/decision` (the package that already owns `ParseRelatedFiles`), and route every existing call site through them: **`… | related-files-single-definition |
| `internal/decision/lint.go`, `internal/decision/lint_test.go` | Remove the warning for empty `## Related Files` in Markdown documents from `kizami lint`. An empty section is now silently accepted. The `## Related Files` sect… | remove-empty-related-files-lint-warning |
| `internal/decision/lint.go`, `internal/decision/lint_test.go` | `kizami lint` の Markdown ドキュメントにおける `## Related Files` 空警告を削除する。 空のセクションは警告なしに受け入れる。`## Related Files` セクション自体はドリフト検出 （`kizami audit`、`kizami blame`、`kizami hoo… | remove-empty-related-files-lint-warning |
| `internal/decision/generate.go`, `internal/decision/decision.go`, `internal/template/template.go`, `cmd/log.go`, `cmd/show.go`, `cmd/status.go` | Remove numeric IDs from document filenames, headings, and metadata. **New filename format**: `YYYY-MM-DD-kebab-case-title.md` **New heading format**: `# Title`… | remove-numeric-ids-from-documents |
| `internal/decision/generate.go`, `internal/decision/decision.go`, `internal/template/template.go`, `cmd/log.go`, `cmd/show.go`, `cmd/status.go` | ドキュメントのファイル名、見出し、メタデータから数値 ID を廃止します。 **新しいファイル名形式**: `YYYY-MM-DD-kebab-case-title.md` **新しい見出し形式**: `# Title`（ID プレフィックスなし） **廃止メタデータ**: `- Supersedes: <slug>`… | remove-numeric-ids-from-documents |
| `go.mod`, `main.go`, `cmd/root.go`, `.goreleaser.yaml`, `CLAUDE.md` | 1. **ツール名を `kizami`（刻み）に変更する。** 設計上の決定やアーキテクチャをコードベースに永続的に刻み込むという コンセプトを表している。コマンド名も `kizami`（`why` を置き換え）とする。 2. **スコープをリビングドキュメント全般に拡張する。** ADR だけでなく、`## Rela… | rename-to-kizami-and-expand-scope |
| `go.mod`, `main.go`, `cmd/root.go`, `.goreleaser.yaml`, `CLAUDE.md` | 1. **Rename the tool to `kizami`** (刻み, Japanese for "to carve/etch"). The name conveys the idea of etching design decisions and architecture into the codebase… | rename-to-kizami-and-expand-scope |
| `internal/initializer/templates/kizami-audit.yml`, `internal/initializer/templates/kizami-promote.yml`, `internal/initializer/init.go`, `internal/decision/audit.go`, `cmd/audit.go` | kizami は2つのスケジュール GitHub Actions ワークフローを生成する。`kizami-audit.yml` は週次で `kizami audit` を実行し、陳腐化した参照が見つかった場合に GitHub Issue を作成する。`kizami-promote.yml` は `main` へのプッシ… | scheduled-ci-workflows |
| `internal/initializer/templates/kizami-audit.yml`, `internal/initializer/templates/kizami-promote.yml`, `internal/initializer/init.go`, `internal/decision/audit.go`, `cmd/audit.go` | kizami generates two scheduled GitHub Actions workflows: `kizami-audit.yml` runs `kizami audit` weekly and opens a GitHub Issue when stale references are found;… | scheduled-ci-workflows |
| `internal/decision/sidecar.go`, `internal/decision/generate.go`, `internal/decision/audit.go`, `internal/search/blame.go` | Introduce `.kizami` sidecar files. A sidecar is a small YAML file placed alongside the managed file, named `<filename>.kizami`. kizami treats the sidecar as the… | sidecar-file-support |
| `internal/decision/sidecar.go`, `internal/decision/generate.go`, `internal/decision/audit.go`, `internal/search/blame.go` | `.kizami` サイドカーファイルを導入します。 サイドカーは管理対象ファイルと同じ場所に置く小さなYAMLファイルで、`<ファイル名>.kizami` という名前にします。 kizamiはサイドカーをドキュメントとして扱い、元ファイルを追跡対象のアーティファクトとして扱います。 サイドカーフォーマット: ```y… | sidecar-file-support |
| `internal/template/template.go` | Reduce the status vocabulary to three values: \| Status \| Meaning \| \| -------------------- \| ---------------------------------- \| \| `Active` \| Currently… | simplify-adr-status |
| `internal/template/template.go` | ステータスを以下の3つに絞る： \| ステータス \| 意味 \| \| -------------------- \| -------------------------------- \| \| `Active` \| 現在有効な判断（デフォルト） \| \| `Inactive` \| 無効。置き換え先なし \|… | simplify-adr-status |
| `go.mod`, `internal/mcp/`, `cmd/mcp.go`, `.mise.toml`, `CONTRIBUTING.md` | 公式SDKを使う。決め手はコード量（自前のstdio JSON-RPC実装自体は実際数百行で済む）ではなく、プロトコル面である：MCPのハンドシェイク、ケーパビリティネゴシエーション、JSON-RPCフレーミングは仕様バージョンをまたいで動き続ける対象であり、SDKの役目はまさにその変動を吸収し、`kizami mcp… | use-official-mcp-go-sdk |
| `go.mod`, `internal/mcp/`, `cmd/mcp.go`, `.mise.toml`, `CONTRIBUTING.md` | Use the official SDK. The deciding factor isn't the code size (a minimal JSON-RPC-over-stdio implementation would indeed be a few hundred lines) but protocol su… | use-official-mcp-go-sdk |

Full text: `kizami show <slug>` · Reverse lookup from changed files: `kizami context <files...>`
<!-- kizami:end -->
