# `kizami_record_decision`: A Gated, New-File-Only Write Tool

- Date: 2026-08-03
- Type: ADR
- Status: Active
- Author: masahiro.kasatani

## Context

Step 4 of the Agent Context Layer ([[agent-context-layer-design]]) is the write half of the roadmap phase: `kizami_record_decision`, an MCP tool letting an agent record a decision immediately after making it, while it still has full context — the moment a human would otherwise have to context-switch back into later, and often just doesn't. This is a materially different risk profile than Steps 1–3 (all read-only): it's the first tool in this phase that writes to the repository, so its safety design needs to be decided deliberately, not inherited from the read tools' ADRs.

## Decision

**Off by default, explicit opt-in.** `kizami mcp` only registers `kizami_record_decision` when started with `--allow-write`. Without the flag, the tool doesn't exist from the agent's point of view — not registered-but-rejecting, genuinely absent, so there's nothing to be tempted to call and no error path to work around.

**New-file-only.** The tool creates exactly one new file under the configured decisions or design directory and touches nothing else — no edits to existing documents, no status updates, no deletions. Concretely, this means **no `supersedes` input**: `cmd/log.go`'s `--supersedes` flag (used by `kizami adr`) also mutates the *old* document's Status line after creating the new one, which is precisely the kind of side-effecting write this tool must not do. An agent that determines an old decision should be superseded still needs a human (or a future, separately-gated tool) to run `kizami status <old> superseded --by <new>`.

**Always `Status: Draft`.** Reuses `decision.CreateFromDraft`/`CreateDesignFromDraft`, which already default to Draft via `template.RenderHeader`. A Draft decision doesn't govern anything yet (`internal/context`'s `governs()` excludes Draft — see [[related-files-single-definition]]), so it can't silently start constraining an agent's own future work until a human reviews and promotes it to Active. It also lands in a PR diff, where the ADR Update Policy's "reviewed like code" story already applies.

**Both document kinds, one tool.** The input takes a `kind` field (`"adr"` default, or `"design"`) plus the section fields for whichever kind, and builds the draft body in the same section order `internal/ai`'s existing AI-draft prompts already use (`## Context`/`## Decision`/`## Consequences`/`## Alternatives Considered` for ADRs; `## Overview`/`## Background`/`## Goals / Non-Goals`/`## Design`/`## Implementation Plan`/`## Open Questions` for design docs) — Related Files is required for both, everything else optional beyond each kind's minimum (`context`+`decision` for ADRs; `overview`+`background`+`design` for design docs).

**No path control from the agent.** The tool takes only a title and section content; the destination directory is always the server's configured `decisions`/`design` directory (from `kizami.toml`, resolved at server startup, identical to every other command). There is no filename or directory input, so there's no path-traversal surface to defend.

## Consequences

- The single largest lever identified in the original problem analysis — decisions aren't written because writing means a costly context switch — gets addressed directly: the agent that just made the call can record it in the same breath, at near-zero marginal cost.
- Every agent-authored decision is a Draft in a diff, reviewed exactly like the code change it accompanies, before it ever governs anything.
- Superseding an old decision remains a two-step, human-involved process (new draft via the tool, then a separate `kizami status` run) rather than one atomic agent action — a deliberate ceiling on this tool's blast radius, not an oversight.
- `--allow-write` is a per-invocation flag, not a config file setting — starting `kizami mcp` with write enabled is a conscious per-session choice each time, not a persisted default someone forgets is on.

## Alternatives Considered

**Allow `supersedes` too, mirroring `kizami adr --supersedes`**
Would close the loop further (an agent could fully retire an old decision itself), but reintroduces exactly the "the tool edits a file the agent didn't tell you it would touch" risk this ADR's new-file-only rule exists to avoid. The two-step manual process is a small ongoing cost for a meaningful reduction in blast radius.

**Default `--allow-write` to on, since Draft status already gates real effect**
Draft status prevents the decision from *governing* anything, but the tool still writes a file to disk unconditionally the moment it's called — an agent could still call it far more often than intended (e.g. a bug causing repeated invocations) or in a context the user didn't expect an MCP server to be writing to their repository. An explicit flag keeps "this session can write to my repo" a decision the human makes, not a default they have to know to turn off.

**One tool per kind (`kizami_record_adr`, `kizami_record_design`)**
Slightly simpler per-tool schema, but doubles the number of write-tool definitions for a distinction (ADR vs. design doc) that's just which section headings apply — better modeled as one field on one tool than two tools an agent has to choose between.

## Related Files

- `cmd/mcp.go`
- `internal/context/record.go`
