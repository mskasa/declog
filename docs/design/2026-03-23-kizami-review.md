# kizami review

- Date: 2026-03-23
- Type: Design
- Status: Active
- Author: masahiro.kasatani

## Overview

`kizami review` surfaces Active documents that have not been updated in a configurable number of months, using git commit history as the source of truth for "last updated" time.

## Background

Living documents are only valuable if they remain accurate. An ADR or design document that describes a decision made two years ago may no longer reflect the current system. Without a way to identify stale documents, teams tend to write documentation once and forget it — leading to misleading references that are worse than no documentation at all.

`kizami review` gives teams a low-friction way to periodically surface documents that may need revisiting, without requiring any metadata beyond what git already tracks.

## Goals / Non-Goals

**Goals:**
- Report Active documents whose last git commit is older than a configurable threshold (default: 6 months)
- Read the threshold from `[review] months_threshold` in `kizami.toml`; allow `--months` flag to override per-run
- Scan all directories listed in `documents.dirs`
- Show the slug/title, last-updated date, and months-since-update for each stale document
- Print a concrete suggestion ("Consider updating, marking as Inactive, or superseding them.")

**Non-Goals:**
- Automatically updating or closing stale documents
- Sending notifications or creating GitHub Issues (that is `kizami audit`'s responsibility)
- Tracking review history within the document itself

## Design

### Staleness Definition

A document is stale if **all** of the following are true:
1. Its `Status` field is `Active` (case-insensitive)
2. The most recent git commit touching its file is older than `now - months`

Draft, Inactive, and Superseded documents are excluded: they are either in-progress or already known to be outdated.

### "Last Updated" via Git

`LastUpdated(path string) (time.Time, error)` runs:

```
git log -1 --format=%ci -- <path>
```

This returns the committer date of the most recent commit touching the file. Using git history rather than filesystem mtime is deliberate: `mtime` changes on checkout and is meaningless for staleness; git commit date reflects when the file was actually modified.

If a file has no git history (e.g. it was never committed), `LastUpdated` returns an error and `FindStale` silently skips that document.

### Threshold Calculation

`FindStale` computes the cutoff as `now.AddDate(0, -months, 0)`. Documents whose `LastUpdated` is before this cutoff are considered stale.

`MonthsAgo(t, now time.Time) int` computes the approximate elapsed months for display, accounting for day-of-month to avoid off-by-one (e.g. April 15 is only 0 months before May 14, not 1).

### Threshold Resolution

Priority (highest to lowest):
1. `--months` flag explicitly passed on the command line
2. `[review] months_threshold` in `kizami.toml` (if > 0)
3. Hard-coded default: `6`

The check `!cmd.Flags().Changed("months")` distinguishes between "flag not passed" and "flag passed with its default value", ensuring the config file takes effect when the flag is omitted.

### Multi-Directory Support

`kizami review` iterates over all directories in `documents.dirs` (resolved by `documentDirs(root, cfg)` in `cmd/root.go`), collects stale documents from each, and reports them together. This covers both `docs/decisions/` and `docs/design/` in a single run.

### Testability

`FindStale` accepts a `lastUpdatedFn func(string) (time.Time, error)` parameter instead of calling `LastUpdated` directly. In tests, a fake function returning controlled timestamps is injected, making the staleness logic testable without a real git repository.

## Open Questions

- **Review timestamp in document metadata**: Currently there is no way to record "this document was reviewed and confirmed accurate on YYYY-MM-DD" without making a git commit. A `Reviewed:` metadata field could extend the staleness signal beyond commit history.
- **Noise for stable decisions**: Some ADRs are deliberately long-lived (e.g., "use Go"). Flagging them as stale every 6 months adds noise. A per-document `review: never` annotation could suppress individual documents from review reporting.

## Related Files

- `internal/decision/review.go`
- `cmd/review.go`
- `internal/config/config.go`
