# 0005: Ripgrep Fallback Strategy

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

`kizami search` needs to perform full-text search across Markdown files.
[ripgrep](https://github.com/BurntSushi/ripgrep) is significantly faster than alternatives and respects `.gitignore`, but it is an external binary that may not be installed on all machines.
Making ripgrep a hard dependency would break `kizami search` for users who have not installed it.

## Decision

Use ripgrep (`rg`) as the primary search backend when it is available on `PATH`, and fall back to a pure Go stdlib implementation (`filepath.Walk` + `strings.Contains`) when it is not.

## Consequences

- `kizami search` works on all machines, even without ripgrep installed
- Users with ripgrep get better performance and `.gitignore` awareness
- The fallback implementation is simpler but sufficient for the typical scale of a `docs/decisions/` directory
- Tests that exercise the ripgrep path must include a skip condition (`t.Skip` when `rg` is not found)

## Alternatives Considered

- **Hard dependency on ripgrep:** Simpler code but breaks for users who have not installed it
- **Pure Go stdlib only:** Maximum portability but slower for large repositories; misses `.gitignore` awareness
- **Embed a search library (e.g., `blevesearch`):** Powerful but significantly increases binary size and complexity for a simple use case

## Related Files

- `internal/search/search.go`
- `internal/search/search_test.go`
