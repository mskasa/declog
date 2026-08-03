# `kizami hook pre-tool-use`: Deterministic Context Injection via Claude Code's PreToolUse Hook

- Date: 2026-08-03
- Type: ADR
- Status: Active
- Author: masahiro.kasatani

## Context

The design doc ([[agent-context-layer-design]]) identifies the manifest (Step 2) and the MCP server (Step 3) as ways to surface governing decisions, but both still depend on the agent's own judgment: reading its context file, or deciding to call a tool. `kizami hook pre-tool-use` is the one deterministic path in this phase — it needs to fire unconditionally before an Edit/Write on a governed file, with no agent judgment involved, closing the gap the other two paths structurally can't.

This requires integrating with Claude Code's actual hook protocol, not a generic notion of "a hook." Checking the current protocol rather than assuming: Claude Code's `PreToolUse` hook receives JSON on stdin (`tool_name`, `tool_input.file_path` for `Edit`/`Write`, `cwd`, etc.) and — critically — can return `{"hookSpecificOutput": {"hookEventName": "PreToolUse", "additionalContext": "..."}}` on stdout with exit code 0 to inject text into Claude's context *without* blocking the tool call. This is exactly the primitive this step needs: `permissionDecision: "deny"`/`"ask"` exist too, but blocking or interrupting the edit was never the goal here — the design doc is explicit that this phase is about surfacing decisions, not gating edits.

## Decision

`kizami hook pre-tool-use` reads the `PreToolUse` JSON from stdin, extracts `tool_input.file_path`, resolves it to a repo-relative path, and calls the same `internal/context.Resolve` every other tool in this phase already uses. If nothing matches, it exits 0 with no output — normal, silent, cheap. If something matches, it prints `additionalContext` and exits 0; it never denies or asks. This tool has exactly one job: make sure the context in Step 1 is impossible to miss at the one moment blocking would be counterproductive to this phase's actual goal.

The injected text is deliberately terser than the agent manifest (Step 2) or an MCP tool response (Step 3): one line per matched decision — slug, one-line title, drift flag if any — with a pointer to `kizami show <slug>` for the body, no decision text inlined. Rationale: the manifest loads once per session and the MCP tools are called on demand, but this hook fires on *every* Edit/Write on a governed file, potentially dozens of times in a session. Even the manifest's already-terse 160-character truncation (see [[agent-manifest-sync-format]]) would compound badly at this call frequency; a pointer is the only shape that stays cheap at that rate.

Failures are silent, mirroring `kizami hook pre-commit`'s existing convention (`return nil` on a git-root or config-load error) — this hook must never be the reason an edit is blocked or a session errors out over something as incidental as "not in a git repo right now."

Wiring is documentation, not code: `kizami init` doesn't manage `.claude/settings.json` in this step. A user who wants this adds it themselves:

```json
{
  "hooks": {
    "PreToolUse": [
      { "matcher": "Edit|Write", "hooks": [{ "type": "command", "command": "kizami hook pre-tool-use" }] }
    ]
  }
}
```

## Consequences

- This is the only path in the whole Agent Context Layer phase that surfaces a decision without any agent cooperation — it fires from Claude Code's own tool-execution loop, before the edit happens.
- Because it's Claude Code-specific (the `PreToolUse` JSON shape, `additionalContext`, and `.claude/settings.json` are not a cross-agent standard the way MCP is), this path only helps in Claude Code specifically, unlike the manifest and MCP server, which work with any MCP-aware or CLAUDE.md/AGENTS.md-reading agent. That asymmetry is accepted, not overlooked: it's the price of the one channel that's actually deterministic today.
- The hook adds one process spawn per Edit/Write on a governed file. Given `internal/context.Resolve`'s cost is bounded by document count (tens to low hundreds, per `internal/config`'s own assumption), this is acceptable, but it's a real per-edit cost the manifest and MCP paths don't have.
- Not wiring `.claude/settings.json` automatically means adoption requires a manual step; this is deferred rather than solved here (see Alternatives).

## Alternatives Considered

**Have `kizami init` write the `.claude/settings.json` hook entry automatically**
Would remove the manual step, but `kizami init` doesn't currently touch anything under `.claude/`, and merging into a JSON file a user may have already customized (existing hooks, other settings) is a meaningfully different and riskier operation than the Markdown/YAML template files `kizami init` writes today. Left as a documented manual step; automatic wiring is a candidate follow-up, not solved speculatively here.

**Include the full decision text in `additionalContext`, matching `kizami context`'s default**
Consistent with Step 1's response shape, but this hook's call frequency (every governed-file edit, not once per query) makes even the summary-only default too expensive here — see Decision. The slug-only pointer is a deliberate, frequency-driven exception to the rest of this phase's "summary by default" rule, not an inconsistency.

**Use `permissionDecision: "ask"` to have the user confirm before proceeding when a decision governs the file**
Would guarantee the context is seen, but turns an informational surfacing mechanism into a gate on every edit of a governed file — directly contradicts the design doc's framing of this step as "surfacing decisions, not gating edits," and would make governed files measurably more annoying to touch, defeating the goal of low-friction adoption.

## Related Files

- `cmd/hook.go`
- `internal/decision/hook.go`
