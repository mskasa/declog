# Unify Related Files Matching Behind a Single Definition

- Date: 2026-08-03
- Type: ADR
- Status: Active
- Author: masahiro.kasatani

## Context

kizami currently computes "does a Related Files entry match a given file" in two independent places:

- `search.Blame` (`internal/search/blame.go`): for directory entries (the trailing-slash convention), `blameDirEntries` inlines its own `strings.HasPrefix` check. Separately, `Blame` also does a full-text search of the whole document body for the file path string — a broader "mentions" concept, not a structured match.
- `decision.CheckHook` (`internal/decision/hook.go`): `hookPathMatches` inlines its own exact-path-or-directory-prefix check.

These aren't just two implementations of the same rule — they disagree. `blameDirEntries` only treats an entry as a directory when it ends with `/` (skipping bare entries like `internal/db` entirely); `hookPathMatches` treats *any* entry, trailing slash or not, as matching both exactly and as a directory prefix (confirmed by `hook_test.go`'s `TestCheckHook_DirectoryMatch`, which asserts that a bare `internal/db` entry matches `internal/db/db.go`). A document's Related Files entry can today govern a file for `kizami hook` but not be recognised as covering it by `kizami blame`'s structural check.

Both exist to answer the same question — "is `file` one of this document's Related Files" — with separate, disagreeing implementations. `internal/context` (see [[agent-context-layer-design]]) is about to expose this as an agent-facing API (`kizami context`, eventually `kizami mcp`). An API can't answer "what governs this file?" two different ways depending on which command an agent happens to call.

The documented, user-facing contract (`docs/site/adr-guide.md`: "You can list directories too — kizami will treat all files under them as related") makes no mention of a trailing-slash requirement. `blameDirEntries`'s stricter behavior is therefore not a deliberate, documented convention being preserved — it under-serves the documented feature. `hookPathMatches`'s permissive behavior is the one that actually matches what users were told the feature does. (Confirmed at the test level too: `hook_test.go`'s `TestCheckHook_DirectoryMatch` and `blame_test.go`'s `TestBlameDirEntries_FileEntryIgnored` assert opposite outcomes for the same input — a bare `internal/db`/`database` entry against a nested file.)

Separately, `decision.Audit` (`internal/decision/audit.go`) answers a related but distinct question — "does a Related Files entry still point at something real" — via a direct `os.Stat` per entry. This is not a matching operation (it doesn't take a candidate file), so it is not folded into the same function, but it needs to stay correct once glob entries exist: `os.Stat` on a glob string (e.g. `internal/**/*_test.go`) will never succeed, since no file is literally named that.

The design also calls for glob support (e.g. `internal/**/*_test.go`), which none of the three call sites have today.

## Decision

Add two small shared primitives to `internal/decision` (the package that already owns `ParseRelatedFiles`), and route every existing call site through them:

**`Match(entry, file string) (kind MatchKind, ok bool)`** — the single definition of "does a Related Files entry match a candidate file," resolving the disagreement in Context by adopting `hookPathMatches`'s more permissive rule as canonical (it is a superset: everything the stricter `blameDirEntries` rule matched, it also matches):
1. Contains `*` or `?` → **glob**. Matched segment-by-segment (`/`-delimited) using `path.Match` semantics per segment, with explicit support for a `**` segment matching zero or more path segments (Go's stdlib `path.Match` has no `**`; implemented as a ~20-line recursive segment matcher rather than adding a dependency).
2. Otherwise, a trailing `/` is stripped (cosmetic only) and the entry matches as **exact** if `file == entry`, or as **dir** if `file` has the (slash-stripped) entry as a path-component prefix — regardless of whether the entry itself was written with a trailing slash. (`internal/db` and `internal/db/` behave identically as Related Files entries.)

**`EntryExists(repoRoot, entry string) bool`** — the single definition of "does this entry still point at something real." Delegates to `os.Stat` for exact/dir entries (unchanged behavior); returns `true` unchecked for glob entries, since verifying a glob still matches at least one real file requires a directory walk that is out of scope for this step (tracked as an open question in [[agent-context-layer-design]]).

Call sites updated:
- `decision.CheckHook` — `hookPathMatches` is removed; it now calls `Match`. Behavior for existing (non-glob) entries is unchanged; glob entries become usable in Related Files for the pre-commit hook as a side effect.
- `search.blameDirEntries` — its inline directory check (previously trailing-slash-only) is replaced with `Match`, filtering to `dir`/`glob` kinds only (the `exact` kind is intentionally excluded here, since an exact-path entry is already found by `Blame`'s full-text search — including it too would just be a redundant duplicate, deduplicated downstream but pointless work). This is the one deliberate behavior change in this ADR: `kizami blame` gains the ability to structurally match a bare directory-style entry (no trailing slash, e.g. `internal/db`) the same way `kizami hook` already did, closing the gap identified in Context. No document in this repository currently uses a bare directory entry (checked against `docs/decisions/` and `docs/design/`), so this has no observable effect today.
- `decision.Audit` — its inline `os.Stat` call is replaced with `EntryExists`. No behavior change for today's entries (exact/dir); glob entries, once they exist, are correctly skipped instead of always reported as missing.
- `internal/context.Resolve` (new) — the primary new consumer; uses `Match` to find which decisions govern the queried files, and `EntryExists` to compute each matched decision's drift state.

## Consequences

- One implementation of "does entry X match file Y," reused by `kizami hook`, `kizami blame`'s structured half, `kizami audit`'s existence check, and the new `kizami context`. Fixing a matching bug or extending the entry syntax now touches one function.
- Glob entries in `## Related Files` become meaningful everywhere except `kizami audit`'s drift check (explicitly, not silently — see Open Questions in the design doc), rather than working in some commands and silently failing in others.
- `kizami hook`'s output is unchanged for all currently-passing behavior (`hook_test.go` is the specification this ADR preserves).
- `kizami blame`'s structural directory-entry matching now also recognizes bare (no trailing slash) directory-style entries, matching what `kizami hook` already did and what `docs/site/adr-guide.md` already documented — see Decision. `TestBlameDirEntries_FileEntryIgnored` (which asserted the old, under-documented behavior) is renamed to `TestBlameDirEntries_BareDirEntryMatches` and updated to assert a match. The full-text "mentions" search remains separate from structured matching either way.
- `kizami audit`'s existing exact/dir entry checking is unchanged; it just no longer duplicates the `os.Stat` logic inline.

## Alternatives Considered

**Fold `Blame`'s full-text search into `Match` too (drop the "mentions" vs. "related" distinction)**
Would give `kizami blame` a single code path, but conflates "this document formally links to the file" (drift-checked, safe for an agent to treat as authoritative) with "this document happens to contain the string" (not drift-checked, often coincidental — e.g. a file path mentioned in prose as an example). Agents need the strong signal for `kizami context`; keeping the concepts distinct in code preserves that. `kizami blame` remains a human-facing "everything that might be relevant" search, which is a legitimate, different job.

**Extend `os.Stat`-based existence checking to also verify glob entries (via a directory walk) in this step**
Correct long-term behavior, but it's a meaningfully different operation (existence of *any* match for a pattern, across the whole tree) from the file-matching this ADR is scoped to, and isn't needed until an actual glob entry exists in a real document. Deferred; tracked as an open question in the design doc rather than solved speculatively here.

**Use a third-party glob library (e.g. `doublestar`)**
Would handle `**` more robustly and cover more glob syntax, but the subset actually needed (per-segment `*`/`?` plus a `**` wildcard segment) is small enough to implement directly in ~20 lines, and CLAUDE.md's stated preference is to avoid dependencies where the stdlib-adjacent effort is modest (see `use-go-over-shell-script`).

## Related Files

- `internal/decision/match.go` (new)
- `internal/decision/audit.go`
- `internal/decision/hook.go`
- `internal/search/blame.go`
- `internal/context/` (new)
