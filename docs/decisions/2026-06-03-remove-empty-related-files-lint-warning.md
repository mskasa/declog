# Remove Empty Related Files Warning from kizami lint

- Date: 2026-06-03
- Status: Active
- Author: masahiro.kasatani
- Supersedes: lint-empty-related-files-as-warning

## Context

`kizami lint` previously emitted a `[warn]` when a Markdown document's `## Related Files`
section was empty, as a reminder against accidental omission.

In practice, multiple legitimate cases exist where an empty `## Related Files` section
is the correct and final state — not a work-in-progress:

- An ADR documenting a **feature deletion**: the related source files are intentionally
  removed in the same commit, so there are no surviving files to list.
- An architectural or process document that spans the entire codebase.
- A document written before the implementation exists (draft stage).

As the number of valid empty cases grows, the warning loses precision as a signal for
accidental omission. It becomes noise rather than a useful reminder.

Additionally, `kizami hook pre-commit` already serves as the primary reminder to create
or update documents at commit time, covering the "forgot to fill in Related Files" case
without requiring a lint check.

## Decision

Remove the warning for empty `## Related Files` in Markdown documents from `kizami lint`.

An empty section is now silently accepted. The `## Related Files` section remains required
for drift detection (`kizami audit`, `kizami blame`, `kizami hook`), and documents that
fill it in continue to benefit from all three mechanisms.

Sidecar (`.kizami`) files retain the error behaviour: a sidecar exists solely to annotate
another file, so an empty `related:` list is always a structural defect.

## Consequences

- `kizami lint` no longer reports issues for documents with an empty `## Related Files` section.
- Feature deletion ADRs and other "intentionally empty" documents pass lint cleanly.
- The pre-commit hook remains the primary mechanism for reminding authors to fill in Related Files.
- There is a small increase in risk that Related Files are left empty by accident, but this
  is acceptable given that the hook provides a reminder at commit time.

## Alternatives Considered

**Add an explicit marker to suppress the warning (e.g. `<!-- intentionally empty -->`)**
Preserves the warning for unintentional omissions but requires authors to learn and apply
a new convention. Adds cognitive overhead without a commensurate benefit.

**Keep the warning, document the legitimate empty cases**
Maintains the signal but requires authors to tolerate noise in known-empty cases, and the
signal degrades further as more legitimate cases are discovered.

## Related Files

- internal/decision/lint.go
- internal/decision/lint_test.go
