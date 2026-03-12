# 0004: Plaintext Markdown Only

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

Design decisions need to be stored somewhere. Options range from a database to structured files to plain text.
The storage format affects portability, Git diff quality, and the tooling required to read decisions.
declog is designed for software projects that already use Git.

## Decision

Store all decisions as plain Markdown files under `docs/decisions/`.
No database, no binary format, no additional metadata files.

## Consequences

- Files are human-readable without any tooling — `cat`, any text editor, or GitHub's web UI all work
- Git-friendly: line-level diffs are meaningful, blame works, history is clear
- Portable: decisions travel with the repository and are accessible offline
- No migration needed: decisions written today are readable in any future environment
- Search is delegated to the filesystem and ripgrep rather than a query language

## Alternatives Considered

- **SQLite database:** Enables structured queries but breaks portability and produces binary diffs
- **JSON/YAML files:** Machine-readable but less pleasant to write and read directly
- **Dedicated ADR service (e.g., external SaaS):** Centralized but creates an external dependency and separates decisions from the codebase
