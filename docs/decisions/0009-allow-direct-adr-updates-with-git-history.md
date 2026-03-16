# 0009: Allow Direct Updates to ADRs Leveraging Git History

- Date: 2026-03-16
- Status: Active
- Author: masahiro.kasatani

## Context

The immutability principle for ADRs originates from an era before distributed version control systems. In pre-Git workflows, overwriting a document meant losing its history, so immutability was the only safe approach.

When the `Superseded` pattern is followed strictly, updating a decision requires creating an entirely new ADR that copies and modifies the original content. This makes it harder to understand what actually changed between the old and new decisions — reviewers must diff two separate files manually rather than relying on `git diff`.

With Git, every change to every file is tracked with author, timestamp, and commit message. The full history of an ADR is always recoverable via `git log -- path/to/adr.md`. There is no technical need to sacrifice clarity for the sake of preserving history.

## Decision

ADRs can be updated directly. Git history is the source of truth for what changed and why.

Guidelines:
- When refining or correcting the same decision, update the ADR in place
- Use `git diff` to communicate what changed; use the commit message to explain why
- Commit messages for ADR updates must include both the what and the why (e.g. `docs: update ADR 0003 - increase pool size from 10 to 20 based on load test`)
- When the direction of a decision changes entirely (i.e. a fundamentally different choice is made), create a new ADR and mark the old one as `Superseded by NNNN`

The boundary between "same decision, refined" and "fundamentally new decision" is a judgment call. When in doubt, prefer creating a new ADR to preserve a clean trail.

## Consequences

- ADR content stays accurate over time without accumulating duplicate files
- `git diff` provides a precise, reviewable change record for every update
- Teams can evolve decisions without the overhead of creating a new ADR for every minor correction
- The `Superseded` status retains its meaning for genuine directional changes, keeping design evolution traceable
- Commit messages carry more weight — vague messages like "update adr" are insufficient

## Alternatives Considered

- **Strict immutability (original policy):** Safe and simple but leads to file proliferation and makes it hard to understand incremental evolution of a decision
- **Append-only updates (add a changelog section at the bottom):** Keeps the original content visible but adds noise and is inconsistent with how other source files are maintained

## Related Files

- `CLAUDE.md`
