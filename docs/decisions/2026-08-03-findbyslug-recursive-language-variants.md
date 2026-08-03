# Make Slug Lookup Recursive Across Nested Language Variants

- Date: 2026-08-03
- Type: ADR
- Status: Active
- Author: masahiro.kasatani

## Context

`decision.FindBySlug(dir, slug)` walks `dir` recursively but stops at the first match (`filepath.SkipAll`). This repository's own EN/JA document pairs are nested under the *same* configured directory (e.g. `docs/decisions/ja/` under `docs/decisions/`), not under separate top-level configured directories. `filepath.WalkDir` visits entries in lexical order, so the EN file (`2026-...-slug.md`) is always found before descending into `ja/`, and the walk stops there — the JA counterpart is never even looked at.

This was discovered while building `kizami mcp`'s `kizami_get_decision` tool (see [[mcp-tools-as-questions-not-verbs]]), which needed to return every document matching a slug and found that reusing `FindBySlug` silently dropped the nested variant. The same underlying function backs three existing commands via `cmd.findAllBySlug`/`cmd.findBySlug`:

- `kizami show <slug>` (tolerates multiple matches, prints all) — today only ever shows the EN half of a nested pair, never both.
- `kizami status <slug> <status>` and `kizami supersede <slug> ...` (error on ambiguity, since they're about to mutate a specific file) — today silently mutate *only* the EN document, unaware the JA one even exists. This is worse than a missing feature: it's a data-integrity risk. Running `kizami status use-x active` on a slug with both an EN and JA document silently leaves the JA document's status field to drift out of sync with its EN counterpart, with no error or warning.

## Decision

Add `decision.FindAllBySlug(dir, slug) ([]*Decision, error)`: the same recursive walk as `FindBySlug`, but without the early exit — it collects every match instead of stopping at the first. `FindBySlug` itself is left unchanged, preserving its existing behavior and callers.

Only `cmd.findAllBySlug` (the shared helper backing `show`/`status`/`supersede`) is switched to call the new function, looping over each configured top-level directory and now correctly collecting *all* matches within each one (not just the first). Its ambiguity-detection consumer, `cmd.findBySlug`, therefore now also correctly detects nested-language-variant ambiguity — which is exactly what it was already designed to do for cross-directory collisions; it was just never able to see this particular case.

`cmd/log.go`'s two direct calls to `decision.FindBySlug` (resolving `--supersedes <slug>` when creating a new ADR) are deliberately left untouched. Making that path ambiguity-detecting too would block `--supersedes` entirely for any slug with an EN/JA pair, with no way in the current CLI surface to disambiguate further (the flag takes only a slug, not a path) — a real capability regression, not just a stricter safety check. Left as a known follow-up if `--supersedes` needs the same treatment.

## Consequences

- `kizami show <slug>` now correctly prints every document matching a slug, including nested language variants — closing a real gap where half of a bilingual pair was invisible to the command.
- `kizami status`/`kizami supersede` now correctly *refuse* to proceed when a slug is ambiguous due to a nested variant, surfacing the same "specify which file you mean" error they already print for cross-directory collisions. This is a behavior change: a command that previously silently succeeded (mutating only the EN file) now errors. That's the intended fix — silent partial mutation of a bilingual pair was the actual bug — but it means any workflow relying on the old silent behavior needs to adjust (there is currently no path-based disambiguator on these commands; a future change would need one if this proves disruptive in practice).
- `kizami adr ... --supersedes <slug>` keeps its current behavior (first match wins) for slugs with a nested variant — not fixed here, tracked as a follow-up.

## Alternatives Considered

**Fix `FindBySlug` itself (remove the early exit) instead of adding a new function**
Would fix `cmd/log.go`'s `--supersedes` path too, but `FindBySlug` returns a single `*Decision` — an ambiguous match would need to either pick one silently (the exact bug being fixed) or return an error, which for `--supersedes` would remove the only route to superseding a slug that has an EN/JA pair, with no replacement mechanism. Changing the return type and forcing every caller to handle that is a larger, riskier change than fixing the one path (`show`/`status`/`supersede`) that's demonstrably broken today.

**Leave the bug and only fix it for `kizami_get_decision`'s new code path**
Would have addressed the immediate need but left `kizami status`/`kizami supersede`'s silent-partial-mutation risk in place — a correctness issue independent of anything MCP-related, worth fixing on its own regardless of the Agent Context Layer roadmap.

## Related Files

- `internal/decision/generate.go`
- `cmd/root.go`
