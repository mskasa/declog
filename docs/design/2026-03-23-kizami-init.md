# kizami init

- Date: 2026-03-23
- Type: Design
- Status: Active
- Author: masahiro.kasatani

## Overview

`kizami init` bootstraps a repository for kizami usage in a single interactive command: it creates the decisions directory, generates `kizami.toml`, and optionally installs CI workflows and a pre-commit hook.

## Background

A new user adopting kizami needs to perform several setup steps: create the document directory, write a config file, set up GitHub Actions workflows, and optionally install a git hook. Each of these has non-trivial boilerplate. Without a setup command, the barrier to adoption is higher and early configuration mistakes are common.

`kizami init` reduces this to a single interactive flow that guides the user through each optional component, with sensible defaults and idempotent checks so it is safe to re-run.

## Goals / Non-Goals

**Goals:**
- Create `docs/decisions/` if it does not exist
- Generate `kizami.toml` with all sections and commented defaults
- Optionally install each of the four optional components via `y/n` prompts:
  - ADR check CI workflow (`adr-check.yml`)
  - pre-commit hook
  - weekly audit CI workflow (`adr-audit.yml`)
  - auto-promote workflow (`kizami-promote.yml`)
- Be idempotent: skip (with a warning) any artifact that already exists
- Also write `~/.config/kizami/config.toml` as a global fallback (via separate `kizami init --global`)

**Non-Goals:**
- Non-interactive/silent mode (all prompts are currently required)
- Generating an initial ADR automatically
- Managing updates to existing config or workflows

## Design

### Initialization Flow

`Initializer.Run()` in `internal/initializer/init.go` executes steps sequentially:

```
kizami init
  │
  ├── 1. createDecisionsDir()     — mkdir docs/decisions/ (skip if exists)
  ├── 2. setupWorkflow()          — prompt: adr-check.yml
  ├── 3. setupHook()              — prompt: pre-commit hook
  ├── 4. setupAuditWorkflow()     — prompt: adr-audit.yml
  ├── 5. setupPromoteWorkflow()   — prompt: kizami-promote.yml
  └── 6. setupConfig()            — write kizami.toml (skip if exists)
```

Each step is independent; a failure in one does not skip subsequent steps (except the step itself returns an error). The `y/n` prompts share a single `bufio.Scanner` over `os.Stdin` so that the input stream is consumed correctly.

### Idempotency

Before writing any file, each step calls `os.Stat` on the target path:
- If the file already exists → print a `⚠️` warning and skip
- If it does not exist → create and print a `✅` confirmation

This makes `kizami init` safe to re-run after partial setup (e.g., if the user said `n` to the audit workflow the first time and wants to add it later).

### Generated Artifacts

#### `kizami.toml`

The default config file is embedded as a Go string constant in `init.go`. It covers all config sections with their default values:

```toml
[ai]
model = "claude-sonnet-4-20250514"

[documents]
dirs = ["docs/decisions", "docs/design"]

[decisions]
dir = "docs/decisions"

[design]
dir = "docs/design"

[audit]
dirs = ["docs/decisions", "docs/design"]

[review]
months_threshold = 6

[editor]
command = "code --wait"
```

#### `.github/workflows/adr-check.yml`

Runs on every pull request. Checks whether commits touching source files (non-docs, non-config paths) include a corresponding document change. Designed as a soft reminder rather than a hard gate — it posts a comment or warning but does not block merges by default.

#### `.git/hooks/pre-commit`

Shell script embedded via `//go:embed templates/pre-commit`. Runs `kizami` availability check and prompts the developer to consider creating a decision record before committing. If a pre-commit hook already exists, the script content is printed to stdout so the user can append it manually — overwriting an existing hook would silently break other tooling.

#### `.github/workflows/adr-audit.yml`

Runs `kizami audit` on a weekly schedule (`cron: '0 0 * * 1'`) and on `workflow_dispatch`. If stale references are found, it creates a GitHub Issue tagged `[kizami audit]`, deduplicating so only one such issue is open at a time. See the Audit and Drift Detection design document for details.

#### `.github/workflows/kizami-promote.yml`

Runs on push to `main`. Promotes documents with `Status: Draft` to `Status: Active` automatically. Includes inline comments explaining the promotion logic so teams can customize the trigger or disable it.

### Template Embedding

All workflow and hook templates are embedded at build time using Go's `//go:embed` directive:

```go
//go:embed templates/adr-check.yml
var adrCheckWorkflow string

//go:embed templates/adr-audit.yml
var adrAuditWorkflow string

//go:embed templates/kizami-promote.yml
var promoteWorkflow string

//go:embed templates/pre-commit
var hookScript string
```

Embedding means the binary is self-contained — no external template files are required at runtime. Updating a template requires rebuilding the binary.

### `Initializer` Struct

```go
type Initializer struct {
    Root   string
    Input  io.Reader
    Output io.Writer
}
```

`Input` and `Output` are injected rather than hardcoded to `os.Stdin`/`os.Stdout`, making the initializer fully testable without a real terminal.

## Open Questions

- **Non-interactive mode**: A `--yes` flag to accept all prompts automatically would be useful for CI-based setup scripts. Not yet implemented.
- **Config updates**: If `kizami.toml` already exists but is missing new keys added in a later version, `kizami init` skips it entirely. A future `kizami init --upgrade` could merge new defaults into an existing config.
- **`docs/design/` creation**: `createDecisionsDir` only creates `docs/decisions/`. The `docs/design/` directory is not created by `kizami init` today. This is a known gap.

## Related Files

- `internal/initializer/init.go`
- `internal/initializer/hook.go`
- `internal/initializer/templates/adr-check.yml`
- `internal/initializer/templates/adr-audit.yml`
- `internal/initializer/templates/kizami-promote.yml`
- `internal/initializer/templates/pre-commit`
- `cmd/init.go`
