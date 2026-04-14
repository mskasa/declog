# Accept Arbitrary Filenames via Content-Based Document Detection

- Date: 2026-04-14
- Status: Active
- Author: masahiro.kasatani

## Context

kizami previously recognised only files matching `YYYY-MM-DD-slug.md` or the legacy `NNNN-slug.md` pattern as managed documents. Any file outside these patterns — including common files such as `ARCHITECTURE.md`, `API-SPEC.md`, or a team's pre-existing documentation — was silently ignored by `kizami list`, `kizami audit`, and related commands.

This naming constraint created friction for teams migrating from other documentation systems, as every existing document had to be renamed before kizami could manage it.

## Decision

Recognise any `.md` file as a kizami document if it contains **both** of the following markers:

1. A line beginning with `- Status:` (required front-matter field)
2. A `## Related Files` section heading (required for drift detection)

Both markers must be present. A file containing only one of them is not treated as a kizami document.

Files matching the existing naming convention (`YYYY-MM-DD-*.md`, `NNNN-*.md`) continue to be recognised via filename pattern matching alone (fast path, no file I/O). Files outside the convention are checked by reading their content (slow path).

**Slug for arbitrary filenames**: the filename without the `.md` extension.
For example, `ARCHITECTURE.md` → slug `ARCHITECTURE`.

**Sort order in `kizami list`**: uses the `- Date:` front-matter field when present, falls back to the file's modification time, and places files where both are unavailable at the end of the list.

## Consequences

- Teams can adopt kizami incrementally: existing documents with established names become manageable by adding two markers, without renaming files.
- `kizami list`, `kizami show`, `kizami blame`, `kizami audit`, and `kizami search` all operate on arbitrary-named documents transparently.
- The content scan for non-pattern files introduces additional file I/O on first run. The impact is negligible for typical `documents.dirs` sizes (tens to low hundreds of files).
- The two-marker requirement (both `- Status:` and `## Related Files`) ensures kizami does not accidentally treat unrelated `.md` files (READMEs, changelogs, etc.) as managed documents.

## Alternatives Considered

**Single-marker recognition (either `- Status:` or `## Related Files`)**
More permissive but risks accidentally treating non-kizami documents as managed. The `## Related Files` section is what enables drift detection, so requiring it ensures every recognised document participates in `kizami audit`.

**Opt-in frontmatter flag (e.g. `- kizami: true`)**
Explicit but requires adding a non-standard field to every document. The two-marker approach reuses fields that kizami authors already write.

**Require renaming (keep current constraint)**
Zero implementation cost but makes adoption by existing teams unnecessarily painful.

## Related Files

- `internal/decision/generate.go`
- `internal/decision/decision_test.go`
