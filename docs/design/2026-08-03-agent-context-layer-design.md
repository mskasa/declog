# Agent Context Layer

- Date: 2026-08-03
- Type: Design
- Status: Active
- Author: masahiro.kasatani

## Overview

kizami currently delivers value only when a human actively runs the CLI (`show`, `blame`, `search`). This design moves the center of gravity to "the right decision surfaces automatically at the right moment" — for AI coding agents specifically, since that is where kizami's `## Related Files` linkage and drift detection have the most leverage. It introduces a single resolver package that four consumers build on: a CLI command, a manifest synced into agent-read files, an MCP server, and (as the final step) an agent write-path for recording new decisions.

## Background

Reachability is not the gap: agents that read a repository can already open `docs/decisions/*.md` directly. The gap is in two other places:

1. **Decisions aren't reliably written at the moment they're made.** Writing an ADR today means a human context-switches at the end of an implementation session — the point of lowest willingness. This is why adoption has not produced the expected payoff: too few decisions accumulate.
2. **Decisions aren't reliably surfaced at the moment they matter.** Even when a document exists, nothing prompts an agent to read it before it edits a governed file. `kizami show`/`blame` require a human (or an agent) to think to run them.

A third, narrower gap: `kizami audit` and `kizami blame` currently compute "relatedness" two different ways —
`search.Blame` (`internal/search/blame.go`) does a full-text search for the file path anywhere in the document, while
`decision.CheckHook` (`internal/decision/hook.go`) structurally parses the `## Related Files` list. An agent-facing API
cannot answer "what governs this file?" with two different definitions of "governs." This has to be unified before it's exposed externally, which is why it is Step 1 of this design (see [[related-files-single-definition]]).

## Goals / Non-Goals

**Goals:**
- One authoritative definition of "which decisions govern a given file," reused by every consumer below (CLI, docs, hook, MCP).
- Reduce the cost of writing a decision close to zero for the case where an agent already knows the answer (it just made the decision).
- Make the *existence* of decisions relevant to a file impossible for an agent to miss, without relying solely on the agent choosing to look.
- Keep each consumer's token/output cost bounded and predictable — this only works in practice if it doesn't crowd out the agent's actual task.

**Non-Goals:**
- Semantic drift detection (whether code still matches the decision's *content*, not just whether the file still exists). Tracked separately in the roadmap; the resolver's `drift` field in this phase is limited to existence checks (what `kizami audit` already does).
- Churn-based staleness scoring for `kizami review`. Separate roadmap item; independent of this design.
- Any LLM/network call inside `kizami mcp`'s read path. The resolver only reads local Markdown/YAML; this keeps the read path free (see the earlier cost discussion) and dependency-free.

## Design

```
                    ┌──────────────────────────────────┐
   kizami context ──┤                                  │
   (CLI / CI)       │   internal/context (resolver)     │
                     │                                  │
   agents sync   ────┤   files[] -> governing decisions  │
   (Step 2)          │   + drift state (existence-only)  │
                      │   single "related" definition    │
   kizami mcp     ───┤                                  │
   (Step 3/4)         └──────────────────────────────────┘
```

`internal/context` is the single new package this phase is built on. It supersedes the "related files" logic
duplicated across `search.Blame` and `decision.CheckHook` (unified per [[related-files-single-definition]]) and adds:

- Matching rules: exact path, directory prefix (existing `trailing-slash/` convention), and glob (new).
- A JSON-serializable result type carrying, per matched decision: slug, title, status, the `## Decision` section only
  (not full body — see token-budget rationale below), which rule(s) matched, and existence-based drift state.
- Superseded decisions are *not* dropped. A decision with `Status: Superseded by <slug>` that still matches is
  returned with a `supersededBy` field, because "this used to be true, now see X" is high-value context for an
  agent that would otherwise repeat a reverted decision.
- `unmatched` files are reported alongside matches — files touched by a change that no decision covers. This is
  inert in Step 1 but is the seed for a later "should this have a decision?" signal.

Four consumers sit on top, each solving a different half of the two gaps in Background:

| Consumer | Command | Gap it addresses | Reliability |
|---|---|---|---|
| CLI / CI | `kizami context <files...> [--json] [--full]` | Neither directly; it's the shared primitive (also powers PR-comment automation) | N/A |
| Agent manifest | `kizami agents sync` | Surfacing — cheap, always-loaded pointer table in CLAUDE.md/AGENTS.md | Depends on the agent reading its own context file |
| MCP server | `kizami mcp` | Surfacing — on-demand, precise | Depends on the agent choosing to call a tool |
| Tool hook | `kizami hook pre-tool-use` | Surfacing — deterministic | Fires unconditionally before Edit/Write, no agent judgment involved |
| Write path | `kizami_record_decision` (MCP) | Writing — collapses the cost to near-zero at the moment an agent has full context | Depends on the agent choosing to call it (mitigated by hook/manifest reminding it) |

Token cost is a first-class design constraint, not an afterthought: `kizami context` and the MCP read tools return only
the `## Decision` section by default (a `full` flag escalates to the whole document). The manifest lists one line per
decision (path -> decision pointer), not decision text. Without this constraint, a 12-decision match on a large diff
would crowd out an agent's actual task context, which is the same failure mode that makes an unused MCP server worse
than no MCP server at all.

## Implementation Plan

1. **Context resolver** (this branch) — `internal/context`, unifying `Blame`/`CheckHook`, adding glob support, and
   `kizami context <files...> [--json] [--full]`.
2. **Agent manifest sync** — `kizami agents sync` (writes the marker-delimited pointer table into CLAUDE.md/AGENTS.md)
   and `kizami agents sync --check` for CI.
3. **MCP server** — `kizami mcp` exposing `kizami_decisions_for_files`, `kizami_search_decisions`, `kizami_get_decision`
   as question-framed tools (not 1:1 CLI-verb mirrors — see the accompanying ADR to be written alongside that step).
4. **Agent-authored decisions** — `kizami_record_decision` (write path, `Status: Draft`, new-file-only, gated behind
   `kizami mcp --allow-write`) and `kizami hook pre-tool-use` (deterministic injection before Edit/Write).

Each step lands as its own PR into `feature/agent-context-layer`; this document's Related Files section is appended
to as each step's files land, per the ADR Update Policy.

## Open Questions

- **Glob implementation.** `path.Match` doesn't support `**`. Resolved for Step 1 in [[related-files-single-definition]].
- **MCP transport/SDK.** Official `github.com/modelcontextprotocol/go-sdk` vs. a small hand-rolled stdio JSON-RPC
  implementation. Deferred to the Step 3 ADR — not needed until the MCP server exists.
- **Token budget defaults.** Whether `## Decision`-only truncation needs a further cap (e.g. max N decisions per
  response) once real multi-decision matches are observed. Revisit after Step 3 ships with real usage.

## Related Files

- `internal/context/`
- `cmd/context.go`
- `cmd/agents.go`
- `internal/decision/match.go`
- `internal/decision/decision.go`
- `internal/decision/audit.go`
- `internal/decision/hook.go`
- `internal/search/blame.go`
- `internal/config/config.go`
