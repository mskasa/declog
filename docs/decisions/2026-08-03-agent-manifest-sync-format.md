# Agent Manifest Format and Sync Strategy for `kizami agents sync`

- Date: 2026-08-03
- Type: ADR
- Status: Active
- Author: masahiro.kasatani

## Context

Step 2 of the Agent Context Layer ([[agent-context-layer-design]]) needs to write a "path → decision" pointer table into files an agent already reads at the start of a session (`CLAUDE.md`, `AGENTS.md`), so decisions are surfaced without the agent needing to think to look for them. This requires several concrete choices the design doc leaves open: how the block is delimited and kept idempotent across repeated runs, what granularity the table uses (and how a decision's summary is kept to one line), which files are targeted by default, and how `--check` fails a CI build. These choices are externally visible (anyone reading the generated Markdown, or scripting against it) and would be a breaking change to reverse later, so they're recorded here rather than decided ad hoc in the implementation.

## Decision

**Delimited, idempotent block.** The generated content is wrapped in HTML comment markers:

```
<!-- kizami:start -->
...
<!-- kizami:end -->
```

`kizami agents sync` replaces everything between an existing marker pair in place; if no marker pair exists in a target file, the block is appended to the end of the file (never inserted at an inferred position — guessing where would risk corrupting hand-written content around it). This makes the command idempotent and safe to run repeatedly, including from a pre-commit hook or CI step.

Marker detection requires the marker string to occupy an entire line by itself (nothing else on that line), not a plain substring search. This was discovered the hard way while dogfooding this feature on kizami's own `CLAUDE.md`: this very ADR's Decision section quotes `<!-- kizami:start -->`/`<!-- kizami:end -->` as example text, and that example ends up embedded mid-line inside a manifest table row once this ADR is itself synced. A substring search matched that decoy occurrence as the block's end, truncating the captured span and making `sync` and `sync --check` disagree with each other immediately after a sync. Requiring the marker to be alone on its line (bounded by line breaks or file start/end) makes it immune to marker-like text appearing inside any decision's own content.

**One row per decision, not per Related Files entry.** Each Active or Superseded decision (see [[agent-context-layer-design]] for why Superseded decisions are kept) contributes exactly one table row, with its Related Files entries joined into a single cell (backtick-wrapped, comma-separated). This matches the token-budget stance established in Step 1 — the whole point of the manifest is to be a cheap, always-loaded pointer, not a second copy of the ADR body.

**Decision summary cells are hard-truncated to ~160 characters, single line.** The summary reuses the same `## Decision`/`## Overview` extraction as `kizami context` (not the full document), with all whitespace/newlines collapsed to single spaces and a `…` suffix if truncated. 160 characters keeps a single row in the same rough per-decision token range already budgeted for `kizami context`'s default response (see the design doc's Design section); a multi-paragraph Decision section would otherwise make the "cheap" manifest expensive to keep loaded at every session start.

**Default targets: `CLAUDE.md` and `AGENTS.md`, whichever exist.** Both are synced if both are present; neither is created if neither exists (`kizami agents sync` errors out with a message to create one first, rather than inventing a project's agent-instructions file on the user's behalf — that decision belongs to the user, not to kizami). Configurable via `kizami.toml`'s new `[agents]` section (`targets = [...]`) for projects using a different filename convention.

**`--check` fails (non-zero exit) if any target's block is missing or stale**, mirroring `kizami lint`'s error/warning exit-code convention (errors fail the build; this is treated as an error, not a warning, since a stale manifest is the exact failure mode this whole phase exists to prevent).

## Consequences

- The manifest can be regenerated and diffed deterministically, which is what makes `--check` possible as a CI gate.
- Hand-written content elsewhere in `CLAUDE.md`/`AGENTS.md` is never touched — only the marked block is replaced.
- A decision with a long, nuanced `## Decision` section is necessarily flattened to a one-line pointer here; the full text remains one `kizami show <slug>` (or `kizami context --full`) away. This is a deliberate lossy compression, not an oversight.
- Projects with unusual agent-file naming need one line of `kizami.toml` config; the common case (CLAUDE.md and/or AGENTS.md at the repo root) needs none.
- Decision summaries are truncated by rune count, not byte count — slicing a Go string by byte index can split a multi-byte UTF-8 character (discovered the same way, mid-implementation: an early version corrupted this repo's own `CLAUDE.md` by cutting a Japanese decision summary mid-character).

## Alternatives Considered

**Insert the block near the top of the file instead of appending**
Would put decisions in front of an agent sooner in a linear read, but requires guessing a safe insertion point in arbitrary existing Markdown, which risks splitting a section or landing inside an unrelated block. Appending is unambiguous; ordering within the file is left to the user to move the markers themselves if they want it earlier (once moved, subsequent syncs still find and replace the same marker pair wherever it is).

**One row per Related Files entry instead of per decision**
Gives a denser, more directly "greppable by path" table, but multiplies row count by the average entries-per-decision and re-introduces exactly the per-entry verbosity Step 1 was designed to keep bounded. A decision with 5 related files would cost 5x the manifest space for the same information.

**No truncation length limit on the summary cell**
Simpler, but a single verbose ADR could balloon the whole manifest's token cost, undermining the reason the manifest exists (see the design doc: "a 12-decision match... would crowd out an agent's actual task context"). A fixed cap keeps the worst case bounded regardless of how any individual ADR is written.

## Related Files

- `internal/context/manifest.go`
- `internal/context/sync.go`
- `cmd/agents.go`
- `internal/config/config.go`
