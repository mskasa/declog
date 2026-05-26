# Treat Empty Related Files as Warning, Not Error, in kizami lint

- Date: 2026-05-26
- Status: Active
- Author: masahiro.kasatani

## Context

`kizami lint` originally reported an error when a document's `## Related Files` section was empty.
In practice, there are legitimate reasons to leave Related Files empty:

- A document is written before the implementation exists (e.g. an ADR drafted at design time).
- An architectural or process document spans the entire codebase rather than specific files.

Blocking CI with an error in these cases creates friction without improving documentation quality.
`kizami audit` and `kizami blame` already skip documents with no Related Files entries,
so the empty state was already a valid no-op at the tooling level.

## Decision

`kizami lint` now emits a `[warn]` line instead of `[error]` when a Markdown document's
`## Related Files` section is missing or empty.
The process exits with code 0 when only warnings are present, so CI is not blocked.

Sidecar (`.kizami`) files retain the error behaviour.
A sidecar exists solely to annotate another file; an empty `related:` list is always a structural defect.

## Consequences

- Documents with empty Related Files pass CI, reducing friction during early authoring stages.
- `kizami audit` and `kizami blame` behaviour is unchanged (they already skipped empty sections).
- Teams that want strict enforcement can treat warnings as errors at the shell level (`kizami lint || exit 1` is unchanged; a wrapper that fails on any output would need adjustment).
- There is a risk that Related Files are left permanently empty after implementation is complete.
  This is a process concern rather than a tooling one; the warning serves as a persistent reminder.

## Alternatives Considered

- **Keep as error, skip Draft status only** — more targeted, but does not cover architectural docs that genuinely have no related files.
- **Add a `--strict` flag** — defers the decision to the caller; adds complexity without a clear default.
- **Remove the check entirely** — loses the reminder value; discarded in favour of the warning approach.

## Related Files

- `internal/decision/lint.go`
- `internal/decision/lint_test.go`
- `cmd/lint.go`
