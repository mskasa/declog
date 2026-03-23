# Simplify ADR Status to Active, Inactive, and Superseded Only

- Date: 2026-03-13
- Status: Active
- Author: masahiro.kasatani

## Context

The original four statuses (`Proposed`, `Accepted`, `Superseded`, `Deprecated`) were inherited from conventional ADR tooling designed for workflows where decisions go through a review and approval cycle.

In declog's workflow, ADRs are committed alongside the code change. This means a decision is already final at the time it is written, making `Proposed` and `Accepted` meaningless placeholders. `Deprecated` does not distinguish between a decision that was simply invalidated and one that was replaced by a newer ADR, making it harder to trace the evolution of design choices.

Two distinct invalidation scenarios need to be represented:
1. A decision is replaced by a new ADR (traceability to the replacement is valuable)
2. A decision becomes invalid with no replacement (e.g. a feature is removed entirely)

## Decision

Reduce the status vocabulary to three values:

| Status               | Meaning                            |
| -------------------- | ---------------------------------- |
| `Active`             | Currently valid decision (default) |
| `Inactive`           | No longer valid; no replacement    |
| `Superseded by NNNN` | Replaced by ADR NNNN               |

The default status is `Active`. ADRs are committed at the same time as the code change, so the decision is considered final at creation.

## Consequences

- The status field is always meaningful; no "placeholder" statuses exist
- `Superseded by NNNN` embeds the replacement ADR ID directly, making design evolution traceable without a separate lookup
- Simpler mental model; every status has a clear, unambiguous meaning
- Existing ADRs (0001–0007) use `Accepted`; these should be treated as equivalent to `Active` and updated opportunistically

## Alternatives Considered

- **Keep the original four statuses:** Compatible with MADR tooling but introduces statuses that are never meaningfully used when ADRs are committed alongside code
- **Two statuses (Active / Inactive):** Simpler, but loses the traceability of which ADR replaced which
- **Free-form status field:** Maximum flexibility but no consistent semantics across ADRs

## Related Files

- `internal/template/template.go`
