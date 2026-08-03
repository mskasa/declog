# Shape `kizami mcp`'s Tools as Questions, Not CLI-Verb Mirrors

- Date: 2026-08-03
- Type: ADR
- Status: Active
- Author: masahiro.kasatani

## Context

With the SDK choice settled ([[use-official-mcp-go-sdk]]), `kizami mcp` needs concrete tool definitions. The most obvious approach — expose `list`/`show`/`search`/`blame` as MCP tools, one per CLI command — was already identified as a trap in the design discussion that motivated this whole roadmap phase: an agent doesn't think in terms of kizami's verbs, it thinks in terms of "what governs the file I'm about to touch" or "has this already been decided." A verb-mirrored tool set also multiplies the number of tool definitions the agent has to hold in context for no corresponding benefit — MCP tool schemas themselves consume tokens at session start, so tool count is not free.

Three further specifics need pinning down, discovered while designing the concrete Go types: what the response should contain by default, how to handle a slug that resolves to more than one document (a real condition in this repo — `agent-context-layer-design` alone already exists as two files, `docs/design/2026-08-03-agent-context-layer-design.md` and its `ja/` counterpart), and what search should return.

## Decision

Three read-only tools for this step (a fourth, write-capable tool is deferred to Step 4 — see [[agent-context-layer-design]]):

**`kizami_decisions_for_files`** — "What governs the files I'm about to read or edit?" Thin wrapper over `internal/context.Resolve`. Input: `{ paths: []string, full?: bool }`. Output: the same `context.Result` JSON shape `kizami context --json` already produces — one schema, two front doors (CLI and MCP), reusing [[related-files-single-definition]]'s matching and the design doc's token-budget stance (summary-only by default, `full` escalates).

**`kizami_search_decisions`** — "Has this already been decided?" Input: `{ query: string, limit?: int, include_superseded?: bool }`. Unlike `kizami search` (which greps every line of every `.md` file and returns raw line matches, including non-kizami Markdown files), this tool enumerates actual kizami documents via `decision.List` first and only then checks each for a case-insensitive match, so results are always real decisions, one row per document, never a stray `README.md`. Each result carries `slug`, `title`, `status`, `decision` (summary), `path`, `superseded_by`, and `excerpt` (the first matching line, for context on *why* it matched). Superseded decisions are excluded by default (`include_superseded` opts back in) — for "has this been decided," the current state is almost always what's wanted; the superseded trail is available on request. `limit` defaults to 10.

**`kizami_get_decision`** — "Give me the full text." Input: `{ slug: string }`. Output: `{ matches: [...] }`, an *array*, not a single document — because a slug is not guaranteed unique across configured document directories. This repository's own EN/JA ADR pairs are a live example: both language versions share a slug and are both legitimate documents.

This intentionally does *not* reuse `decision.FindBySlug` (the function backing `kizami show`). `FindBySlug` stops at the first match within a directory tree (`filepath.SkipAll`) — discovered while building this tool, it means `kizami show <slug>` today only ever returns the EN half of an EN/JA pair nested under the same configured directory (e.g. `docs/decisions/ja/`), never both, despite `cmd.findAllBySlug` looping over configured *top-level* directories expecting multiple results. `kizami_get_decision` is backed by a new `context.GetBySlug`, which scans every recognized document directly and so finds both. This is a real, pre-existing gap in `kizami show`, not something this ADR fixes — flagged for the repository owner to decide on separately.

Tool responses are returned both as a single JSON `TextContent` block (for clients that only render text) and as typed structured content (via the SDK's generic `AddTool[In, Out]`, which derives the output schema from the Go struct) — no hand-maintained JSON Schema documents.

No write capability in this step. `kizami mcp` as shipped here only reads `docs/decisions/`/`docs/design/`; `kizami_record_decision` (Step 4) is a separate, explicitly opt-in tool gated behind `--allow-write`, not bundled here.

## Consequences

- Three tool definitions, not four-plus — smaller context footprint per session than a verb-mirrored set, and each tool answers a question an agent actually has, not a command it has to already know exists.
- `kizami_get_decision` callers must handle an array, even for the common single-match case — a small ergonomic cost paid once, in exchange for never silently picking the wrong language variant or erroring on a legitimately ambiguous slug.
- `internal/context` now has two ways to find documents by slug: the new `GetBySlug` (full scan, finds every match) and the pre-existing `decision.FindBySlug` (first-match, used by `kizami show`/`status`/`supersede`). They are not interchangeable, and `kizami show`'s current behavior for nested EN/JA pairs is arguably a bug worth fixing on its own — out of scope here, but worth a follow-up decision.
- `kizami_search_decisions`'s document-first enumeration is slower on very large repositories than a raw grep would be (it opens every recognized document, not just files a grep engine flags), but bounded by the same "tens to low hundreds of documents" scale `internal/config`'s own doc comment already assumes for this tool's target usage.
- Because there's no write tool yet, `kizami mcp` needs no `--allow-write` flag or write-safety design in this step; that entire concern is deferred intact to Step 4.

## Alternatives Considered

**Mirror the CLI verbs 1:1 (`list`, `show`, `search`, `blame`, `audit` as separate tools)**
This is the exact failure mode the design doc's original analysis warned against: CLI verbs assume a human decided to run a command; an agent's actual need is answered by a smaller set of question-shaped tools sitting on top of the same resolver, not a larger set of verb-shaped ones.

**`kizami_get_decision` errors on an ambiguous slug (matching `kizami status`/`kizami supersede`'s stricter `findBySlug`)**
Consistent with some existing commands, but those commands error because they're about to *mutate* a specific file and picking the wrong one would be destructive. `kizami_get_decision` is read-only; returning all matches costs nothing and is strictly more useful to an agent than an error it would just have to work around by trying again with more context it doesn't have.

**Single combined `kizami_query` tool with a mode parameter instead of three tools**
Fewer tool definitions, but collapses three different response shapes and semantics behind one schema, which a generic-parameter design tends to make harder for a model to use correctly than three narrowly-typed tools — MCP tool-selection quality generally benefits from specific tools with clear names over one overloaded tool with a mode string.

## Related Files

- `cmd/mcp.go`
- `internal/context/search.go`
- `internal/context/get.go`
