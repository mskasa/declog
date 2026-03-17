# 0010: Rename to kizami and Expand Scope to Living Documentation

- Date: 2026-03-17
- Status: Active
- Author: masahiro.kasatani
- Supersedes: 0006

## Context

declog was originally designed as an ADR (Architecture Decision Record) tool.
The core value proposition was "record the reasoning behind decisions."

However, the deeper problem in software development teams is broader:
**design documents (architecture docs, API specs, detailed design docs) consistently
drift from the actual code over time.** In every project the author has worked on,
design documentation was out of sync with the codebase — 100% of the time.

`why audit` introduced a mechanism to detect this drift by checking whether files
listed in the `## Related Files` section of a document still exist in the repository.
This mechanism is not specific to ADRs — it applies to any Markdown document that
references source files.

This led to a realization: the tool's core value is not "ADR management" but
**"Git-native drift detection for living documentation."** ADRs are one type of
living document, but the same approach works for design documents, API specs,
architecture diagrams references, and more.

The existing tool name (`declog` / `why`) communicates "decision log," which is
too narrow for the expanded scope and makes the tool feel ADR-specific to new users.

## Decision

1. **Rename the tool to `kizami`** (刻み, Japanese for "to carve/etch").
   The name conveys the idea of etching design decisions and architecture
   into the codebase permanently alongside the code itself.
   The command name becomes `kizami` (replacing `why`).

2. **Expand scope to living documentation.**
   The tool will support any Markdown document with a `## Related Files` section,
   not just ADRs. New document types (design docs, API specs, etc.) will be
   supported via a `--type` flag on `kizami log`.

3. **Generalize `kizami audit`** to scan multiple configurable directories,
   not just `docs/decisions/`.

The `## Related Files` mechanism and Git-native philosophy are preserved as-is.

The tool addresses document drift through two complementary mechanisms:
- **Proactive:** the pre-commit hook prompts developers to create or update documents
  at the moment of committing code changes
- **Reactive:** `kizami audit` detects drift periodically (or in CI) by verifying
  that files listed in `## Related Files` sections still exist in the repository

## Consequences

- The tool gains a unique, internet-searchable name with no naming conflicts
- The expanded scope addresses a pain point felt in virtually every software project
- ADR functionality is fully preserved; existing ADR users have a migration path
- `kizami audit` becomes the core differentiator: lightweight, Git-native drift detection
  that requires no infrastructure beyond a text editor and Git
- The command name change (`why` → `kizami`) is a breaking change for existing users,
  but the tool is pre-public-release so the impact is minimal
- The Japanese name gives the tool a distinctive identity in the international OSS ecosystem

## Alternatives Considered

- **Keep the name `declog` / `why` and expand scope:** The name would contradict the
  expanded purpose and confuse new users coming for design document management
- **Create a separate tool for design docs:** Would duplicate the core audit mechanism
  and split maintenance effort
- **Rename to `docrift` or `driftless`:** Descriptive but less unique and memorable
  than `kizami`

## Related Files

- `go.mod`
- `main.go`
- `cmd/root.go`
- `.goreleaser.yaml`
- `CLAUDE.md`
