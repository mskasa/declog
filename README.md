# declog

**`why`** — A minimal CLI tool to record and search architectural design decisions.

[日本語版 README はこちら](README.ja.md)

---

![demo](docs/demo.gif)

---

Design decisions tend to get scattered across Issues, PRs, and Slack — and eventually lost.
`why` saves them as Markdown files alongside your code, so the reasoning behind every choice stays in the repository forever.

```
$ why log "use PostgreSQL over SQLite"
Created: docs/decisions/0007-use-postgresql-over-sqlite.md

$ why list
ID    Date        Status    Title
--    ----        ------    -----
0007  2026-03-12  Proposed  use PostgreSQL over SQLite
0006  2026-03-12  Accepted  Command Name "why"
...

$ why search "PostgreSQL"
docs/decisions/0007-use-postgresql-over-sqlite.md:1: # 0007: use PostgreSQL over SQLite
```

## Installation

### go install (recommended if you have Go)

```bash
go install github.com/mskasa/declog@latest
```

### Download binary

Download the latest binary for your platform from the [Releases page](https://github.com/mskasa/declog/releases).

**macOS / Linux**

```bash
# macOS (Apple Silicon)
curl -L https://github.com/mskasa/declog/releases/latest/download/why_darwin_arm64.tar.gz | tar xz
mv why /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/mskasa/declog/releases/latest/download/why_darwin_amd64.tar.gz | tar xz
mv why /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/mskasa/declog/releases/latest/download/why_linux_amd64.tar.gz | tar xz
mv why /usr/local/bin/
```

**Windows (PowerShell, requires administrator)**

```powershell
# amd64
Invoke-WebRequest -Uri https://github.com/mskasa/declog/releases/latest/download/why_windows_amd64.zip -OutFile why.zip
Expand-Archive why.zip -DestinationPath why_bin
Move-Item why_bin\why.exe C:\Windows\System32\why.exe
Remove-Item why.zip, why_bin -Recurse
```

## Quick Start

```bash
# 1. Record a decision
why log "use PostgreSQL over SQLite"
# Opens the generated Markdown file in your $EDITOR

# 2. List all decisions
why list

# 3. View a specific decision
why show 7

# 4. Search by keyword
why search "PostgreSQL"

# 5. Update a status
why status 7 accepted
why status 3 superseded --by 7
```

## Commands

| Command | Description |
|---|---|
| `why log "<title>"` | Create a new decision record and open it in `$EDITOR` |
| `why list` | List all decisions in reverse chronological order |
| `why show <id>` | Print the full content of a decision |
| `why search <keyword>` | Search decisions by keyword |
| `why status <id> <status>` | Update the status of a decision |

### Statuses

| Status | Meaning |
|---|---|
| `Proposed` | Under consideration (default) |
| `Accepted` | Approved and adopted |
| `Superseded` | Replaced by another decision |
| `Deprecated` | No longer applicable |

### `why status` examples

```bash
why status 3 accepted
why status 3 superseded --by 5   # marks 0003 as superseded by 0005
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
- Status: Proposed
- Author: you

## Context

<!-- Why this decision was needed -->

## Decision

<!-- What was decided -->

## Consequences

<!-- Impact, benefits, trade-offs -->

## Alternatives Considered

<!-- Options not adopted, and why -->
```

## Search

`why search` uses [ripgrep](https://github.com/BurntSushi/ripgrep) when available for speed, and falls back to a pure Go implementation when it is not installed — so it works everywhere.

## Design Decisions

This repository uses `why` to record its own design decisions. Browse [`docs/decisions/`](docs/decisions/) to see it in action.

## License

MIT
