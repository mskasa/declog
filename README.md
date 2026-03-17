# kizami

<p align="center">
  <img src="docs/assets/logo.svg" alt="kizami" width="400">
</p>

**`kizami`** — A minimal CLI tool to maintain living documentation alongside code, with automatic drift detection.

[日本語版 README はこちら](README.ja.md)

---

Design decisions tend to get scattered across Issues, PRs, and Slack — and eventually lost.
`kizami` saves them as Markdown files alongside your code, so the reasoning behind every choice stays in the repository forever.

```
$ kizami log "use PostgreSQL over SQLite"
Created: docs/decisions/0007-use-postgresql-over-sqlite.md

$ kizami list
ID    Date        Status    Title
--    ----        ------    -----
0007  2026-03-12  Active    use PostgreSQL over SQLite
0006  2026-03-12  Active    Command Name "kizami"
...

$ kizami search "PostgreSQL"
docs/decisions/0007-use-postgresql-over-sqlite.md:1: # 0007: use PostgreSQL over SQLite
```

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

# 2. Record a decision
kizami log "use PostgreSQL over SQLite"
# Opens the generated Markdown file in your $EDITOR

# 3. List all decisions
kizami list

# 4. View a specific decision
kizami show 7

# 5. Search by keyword
kizami search "PostgreSQL"

# 6. Update a status
kizami status 7 inactive
kizami status 3 superseded --by 7
```

## Commands

| Command | Description |
|---|---|
| `kizami init` | Initialize the decisions directory and optional GitHub Actions workflow |
| `kizami log "<title>"` | Create a new decision record and open it in `$EDITOR` |
| `kizami list` | List all decisions in reverse chronological order |
| `kizami show <id>` | Print the full content of a decision |
| `kizami search <keyword>` | Search decisions by keyword |
| `kizami status <id> <status>` | Update the status of a decision |
| `kizami blame <file>` | Find decisions that reference a given file |
| `kizami audit` | Detect drift between Related Files sections and actual code |
| `kizami review` | Detect long-stale decisions |

### Statuses

| Status | Meaning |
|---|---|
| `Active` | Currently valid decision (default) |
| `Inactive` | No longer applicable, no replacement |
| `Superseded by NNNN` | Replaced by another decision |

### `kizami status` examples

```bash
kizami status 3 inactive
kizami status 3 superseded --by 5   # marks 0003 as superseded by 0005
```

## Decision File Format

Decisions are saved as Markdown files under `docs/decisions/` using a [MADR](https://adr.github.io/madr/)-compatible template:

```
docs/decisions/
├── 0001-use-go-over-shell-script.md
├── 0002-use-cobra-for-cli-framework.md
└── ...
```

File naming: `NNNN-kebab-case-title.md` — the number is auto-incremented.

```markdown
# 0007: Use PostgreSQL over SQLite

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
