# Sidecar File Support for Non-Markdown Documents

- Date: 2026-04-17
- Status: Active
- Author: masahiro.kasatani

## Context

kizami previously only managed Markdown documents (`.md` files with `- Status:` and `## Related Files` markers).
This meant non-Markdown files such as CSV test matrices, OpenAPI specs, SQL schemas, and images could only be tracked indirectly — by listing them in a separate Markdown document's `## Related Files` section.

That indirect approach required creating a Markdown document solely to act as a bridge, which added friction.
The goal is to allow any file type to become a first-class kizami artifact without altering the file itself.

## Decision

Introduce `.kizami` sidecar files.
A sidecar is a small YAML file placed alongside the managed file, named `<filename>.kizami`.
kizami treats the sidecar as the document and the original file as the artifact being tracked.

Sidecar format:

```yaml
title: Test matrix for user flow
date: 2026-04-17
author: masahiro.kasatani
related:
  - tests/user_flow_test.go
```

Key design choices:

- **No `status` field**: sidecars represent factual relationships, not decisions under review.
  They are always treated as Active and are always included in `kizami audit`.
- **`date` = creation date**: consistent with Markdown ADRs. Update history is tracked via git log.
- **No external YAML library**: the format is simple enough to parse line-by-line, keeping zero new dependencies.
- **Slug = managed filename**: for `test_matrix.csv.kizami` the slug is `test_matrix.csv`,
  so `kizami show test_matrix.csv` finds it naturally.

## Consequences

- Any file type can now be managed by kizami with minimal overhead (one small sidecar file).
- `kizami blame`, `kizami audit`, `kizami list`, and `kizami show` all support sidecar files automatically.
- The sidecar format is intentionally minimal. Reasoning and context belong in a linked Markdown ADR or design doc.

## Alternatives Considered

- **CSV comment metadata** (`# kizami:related: ...`): intrusive, format-specific, and requires per-format parsers.
- **Richer sidecar with description field**: adds value but blurs the line with Markdown docs; deferred.

## Related Files

- internal/decision/sidecar.go
- internal/decision/generate.go
- internal/decision/audit.go
- internal/search/blame.go
