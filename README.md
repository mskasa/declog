# kizami

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./docs/assets/kizami-logo-dark.svg">
    <img src="./docs/assets/kizami-logo-light.svg" alt="kizami logo" width="400">
  </picture>
</p>

**`kizami`** — A minimal CLI tool to maintain living documentation alongside code, with automatic drift detection.

[![CI](https://github.com/mskasa/kizami/actions/workflows/ci.yml/badge.svg)](https://github.com/mskasa/kizami/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mskasa/kizami)](https://goreportcard.com/report/github.com/mskasa/kizami)
[![GitHub release](https://img.shields.io/github/v/release/mskasa/kizami)](https://github.com/mskasa/kizami/releases/latest)

[日本語版 README はこちら](README.ja.md)

---

Design decisions tend to get scattered across Issues, PRs, and Slack — and eventually lost.
`kizami` saves them as Markdown files alongside your code, so the reasoning behind every choice is recorded and kept accurate over time.

```
$ kizami adr "use PostgreSQL over SQLite"
Created: docs/decisions/2026-03-12-use-postgresql-over-sqlite.md

$ kizami list
Slug                            Date        Status    Title
----                            ----        ------    -----
use-postgresql-over-sqlite      2026-03-12  Draft     use PostgreSQL over SQLite
command-name-kizami             2026-03-12  Active    Command name "kizami"
...

$ kizami search "PostgreSQL"
docs/decisions/2026-03-12-use-postgresql-over-sqlite.md:1: # use PostgreSQL over SQLite
```

## Documentation

Full documentation is available at **[mskasa.github.io/kizami](https://mskasa.github.io/kizami/)** — workflow guides, ADR writing tips, configuration reference, and more.

## Installation

### go install (recommended if you have Go)

```bash
go install github.com/mskasa/kizami@latest
```

### Download binary

Download the latest binary for your platform from the [Releases page](https://github.com/mskasa/kizami/releases).

**macOS / Linux**

```bash
# macOS (Apple Silicon)
curl -L https://github.com/mskasa/kizami/releases/latest/download/kizami_darwin_arm64.tar.gz | tar xz
mv kizami /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/mskasa/kizami/releases/latest/download/kizami_darwin_amd64.tar.gz | tar xz
mv kizami /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/mskasa/kizami/releases/latest/download/kizami_linux_amd64.tar.gz | tar xz
mv kizami /usr/local/bin/
```

**Windows (PowerShell, requires administrator)**

```powershell
# amd64
Invoke-WebRequest -Uri https://github.com/mskasa/kizami/releases/latest/download/kizami_windows_amd64.zip -OutFile kizami.zip
Expand-Archive kizami.zip -DestinationPath kizami_bin
Move-Item kizami_bin\kizami.exe C:\Windows\System32\kizami.exe
Remove-Item kizami.zip, kizami_bin -Recurse
```

## Quick Start

```bash
# 1. Initialize your decisions directory
kizami init

# 2. Record an architectural decision (ADR)
kizami adr "use PostgreSQL over SQLite"
# Opens the generated Markdown file in your $EDITOR

# 2b. Create a design document
kizami design "connection pool design"

# 3. List all decisions
kizami list

# 4. View a specific decision
kizami show use-postgresql-over-sqlite

# 5. Search by keyword
kizami search "PostgreSQL"

# 6. Update a status
kizami status use-postgresql-over-sqlite inactive
kizami status use-sqlite superseded --by use-postgresql-over-sqlite
```

## Use Cases

### Onboarding new team members

Instead of asking a senior engineer "why is the code structured this way?", new members can look it up themselves.

```bash
$ kizami list
Slug                            Date        Status    Title
----                            ----        ------    -----
use-postgresql-over-sqlite      2025-09-01  Active    Use PostgreSQL over SQLite
auth-jwt-strategy               2025-10-15  Active    Auth JWT strategy
connection-pool-design          2025-11-20  Active    Connection pool design
...

$ kizami show use-postgresql-over-sqlite
# Use PostgreSQL over SQLite
- Date: 2025-09-01
- Status: Active

## Context
Load tests at 500 concurrent users revealed SQLite's write lock as a bottleneck.

## Decision
Switch to PostgreSQL. Accepted the operational overhead in exchange for write scalability.
```

Even a new team member can understand the reasoning behind past decisions right alongside the code.

### Answering "why?" in code reviews

When a reviewer asks about a design choice, point to the ADR instead of re-explaining from scratch.

```bash
$ kizami blame internal/db/db.go
docs/decisions/2025-09-01-use-postgresql-over-sqlite.md
```

Reviewers get the context instantly, and discussions stay focused on what matters.

### Catching stale documentation

`kizami review` surfaces documents that haven't been updated in a long time, so outdated docs don't get quietly ignored.

```bash
$ kizami review
docs/decisions/2025-08-20-redis-caching-strategy.md
  Last updated 9 months ago — worth revisiting?
```

### Providing context to AI coding assistants

Because decisions are stored as plain Markdown files, you can feed past decisions directly to an AI assistant.

```bash
$ kizami show use-postgresql-over-sqlite
# Use PostgreSQL over SQLite
- Status: Active

## Context
Load tests at 500 concurrent users revealed SQLite's write lock as a bottleneck.

## Decision
Switch to PostgreSQL. Accepted the operational overhead in exchange for write scalability.
```

Pass this to your AI assistant and it can make suggestions that respect your existing constraints and trade-offs — no need to re-explain the background every time.

## Commands

| Command | Description |
|---|---|
| `kizami init` | Initialize the decisions directory and optional GitHub Actions workflows |
| `kizami adr "<title>"` | Create a new ADR and open it in `$EDITOR` |
| `kizami design "<title>"` | Create a new design document and open it in `$EDITOR` |
| `kizami list` | List all documents in reverse chronological order |
| `kizami show <slug>` | Print the full content of a document |
| `kizami search <keyword>` | Search documents by keyword |
| `kizami status <slug> <status>` | Update the status of a document |
| `kizami supersede <slug> "<title>"` | Supersede an existing document and create a new one |
| `kizami blame <file>` | Find documents that reference a given file |
| `kizami context <files...>` | Show which documents govern the given files, with drift status (`--json`, `--full`) |
| `kizami agents sync` | Sync a decisions pointer table into CLAUDE.md / AGENTS.md (`--check` for CI) |
| `kizami audit` | Detect drift between Related Files sections and actual code |
| `kizami lint` | Validate document structure for CI |
| `kizami review` | Detect long-stale documents |

### Statuses

| Status | Meaning |
|---|---|
| `Active` | Currently valid decision (default) |
| `Inactive` | No longer applicable, no replacement |
| `Superseded by <slug>` | Replaced by another document |

### `kizami status` examples

```bash
kizami status use-sqlite inactive
kizami status use-sqlite superseded --by use-postgresql-over-sqlite
```

## Decision File Format

Decisions are saved as Markdown files under `docs/decisions/` using a [MADR](https://adr.github.io/madr/)-compatible template:

```
docs/decisions/
├── 2026-03-12-use-go-over-shell-script.md
├── 2026-03-12-use-cobra-for-cli-framework.md
└── ...
```

File naming: `YYYY-MM-DD-kebab-case-title.md`

```markdown
# Use PostgreSQL over SQLite

- Date: 2026-03-12
- Status: Active
- Author: you

## Context

<!-- Why this decision was needed -->

## Decision

<!-- What was decided -->

## Consequences

<!-- Impact, benefits, trade-offs -->

## Alternatives Considered

<!-- Options not adopted, and why -->

## Related Files

<!-- List files related to this decision (e.g. internal/db/db.go) -->
```

## Drift Detection

The `## Related Files` section links any document to the source files it references.
`kizami audit` detects when those files are deleted or moved — keeping documentation honest.

```bash
kizami audit
# Checks all Related Files entries in docs/decisions/ and reports missing files
```

## Search

`kizami search` uses [ripgrep](https://github.com/BurntSushi/ripgrep) when available for speed, and falls back to a pure Go implementation when it is not installed — so it works everywhere.

## Design Decisions

This repository uses `kizami` to record its own design decisions. Browse [`docs/decisions/`](docs/decisions/) to see it in action.

## License

MIT
