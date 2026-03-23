# 0011: Remove Numeric IDs from Documents

- Date: 2026-03-23
- Status: Draft
- Author: masahiro.kasatani

## Context

Documents created by kizami currently use a 4-digit numeric ID prefix in their filenames (e.g., `0001-use-go-over-shell-script.md`). This ID is embedded in the filename, the document heading (`# 0001: Title`), and referenced in supersedence metadata (`- Supersedes: 0003`, `Status: Superseded by 0007`).

This scheme was adopted for short CLI references (`kizami show 3`) and to express supersedence relationships compactly.

However, the ID-based approach introduces significant structural constraints:

- **Subdirectory organisation is broken**: `NextID` scans only the immediate directory. Placing documents in subdirectories (e.g., `docs/decisions/2024/`, `docs/decisions/ja/`) would cause ID collisions or silently miss files.
- **Incremental complexity**: `NextID`, `FindByID`, and `List` all contain ID-specific logic with special-casing for the 4-digit format.
- **Incompatibility with MADR convention**: The broader ADR ecosystem (adr-tools, MADR) uses date- or slug-based filenames without numeric IDs.
- **Fragile references**: Supersedence recorded as `Superseded by 0007` is opaque without looking up the file; a filename slug is self-documenting.

## Decision

Remove numeric IDs from document filenames, headings, and metadata.

**New filename format**: `YYYY-MM-DD-kebab-case-title.md`
**New heading format**: `# Title` (no ID prefix)
**Supersedence metadata**: `- Supersedes: <slug>` (filename without extension)

The date prefix preserves chronological sort order. The slug provides a stable, human-readable identifier that can be used in CLI commands and cross-references.

CLI commands that currently take a numeric ID will accept the slug instead:

- `kizami show <slug>`
- `kizami status <slug> <status>`
- `kizami supersede <slug> <slug>`

## Consequences

**Benefits:**
- Subdirectory organisation (by year, language, team, etc.) becomes possible without ID collision
- `NextID`, `FindByID`, and related numbering logic can be removed, simplifying the codebase
- Filenames are self-documenting; no need to look up `0007` to understand the reference
- Closer alignment with MADR and adr-tools conventions

**Trade-offs:**
- Breaking change: existing documents with `NNNN-` prefixes must be migrated (rename files, update headings and cross-references)
- `kizami show 3` shorthand is lost; callers must use the slug
- Slug-based lookup requires prefix or exact matching instead of integer comparison

**Migration path:**
- Provide a `kizami migrate` subcommand (or script) to rename existing files and update internal references in bulk
- During a transition period, the parser will accept both `NNNN-slug.md` and `YYYY-MM-DD-slug.md` formats to avoid breaking existing repositories

## Alternatives Considered

**Keep IDs, fix subdirectory support with a global ID registry**
Would require a lock file or index to coordinate IDs across subdirectories. Adds complexity without removing the root cause.

**Keep IDs, restrict documents to a flat directory**
Prevents legitimate organisational needs (per-language directories, year-based archives). Inconsistent with how other tools handle growing document sets.

**Use UUIDs instead of sequential IDs**
Collision-free but completely opaque and incompatible with human references in commit messages and PR descriptions.

## Related Files

- `internal/decision/generate.go`
- `internal/decision/decision.go`
- `internal/template/template.go`
- `cmd/log.go`
- `cmd/show.go`
- `cmd/status.go`
