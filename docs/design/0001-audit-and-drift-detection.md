# 0001: Audit and Drift Detection

- Date: 2026-03-23
- Type: Design
- Status: Draft
- Author: masahiro.kasatani

## Overview

`kizami audit` detects when source files listed in a document's `## Related Files` section no longer exist in the repository.
This keeps living documents honest by surfacing stale references automatically, rather than relying on developers to remember to update them.

## Background

The core value of kizami is maintaining a traceable link between decisions/designs and the code they describe.
That link is expressed through the `## Related Files` section in every document.

Without automated checking, this link degrades silently:
- A file is renamed → the document still points to the old path
- A module is deleted → the document describes code that no longer exists
- A directory is reorganized → the document's references become wrong

Developers rarely update documentation when they move files.
`kizami audit` makes this drift visible before it becomes permanent.

## Goals / Non-Goals

**Goals:**
- Detect file paths and directory paths in `## Related Files` that no longer exist
- Run both as a CLI command (`kizami audit`) and as a scheduled CI job
- Support multiple document directories (ADRs + design docs)
- Skip Draft documents (they are in-progress and not yet authoritative)

**Non-Goals:**
- Symbol-level drift detection (e.g., function was renamed inside a file) — file existence is the check boundary
- Auto-updating stale references — audit only reports; humans decide the fix
- Checking documents outside the configured `audit.dirs`

## Design

### The Related Files Mechanism

Every document created by kizami contains a `## Related Files` section:

```markdown
## Related Files

- `internal/decision/audit.go`
- `internal/search/blame.go`
- `cmd/audit.go`
```

This section is the authoritative source of the document–code link.
It is parsed by `ParseRelatedFiles` (`internal/decision/audit.go`) line by line:
1. Scan until `## Related Files` heading is found
2. Collect list items (`- path` or `- \`path\``)
3. Stop at the next `##` heading

### Directory Prefix Entries

A path ending with `/` is treated as a directory entry and matches any file under that path:

```markdown
## Related Files

- `internal/search/`
```

This means: "this document is related to everything under `internal/search/`."

- **For `kizami audit`**: the directory path itself is checked with `os.Stat`. If `internal/search/` is deleted entirely, the audit reports it as missing.
- **For `kizami blame <file>`**: `blame` additionally matches directory entries. A query for `internal/search/search.go` will surface documents listing `internal/search/`. See `blameDirEntries` in `internal/search/blame.go`.

This convention scales well: a document about a subsystem can reference the whole directory instead of listing every file individually.

### Drift Detection Algorithm

`Audit(dir, repoRoot string)` in `internal/decision/audit.go`:

```
1. List all documents in dir (sorted by ID)
2. For each document:
   a. Skip if Status != "Active" (case-insensitive)
   b. Parse Related Files entries
   c. Skip if no entries
   d. For each entry: os.Stat(filepath.Join(repoRoot, entry))
   e. Collect entries where os.IsNotExist(err) == true
3. Return AuditResult{Decision, MissingFiles} for each document with ≥1 missing path
```

**Why only Active documents?**
Draft documents are works-in-progress. Their Related Files may be aspirational (files not yet created) or exploratory. Auditing them would generate noise. Only Active (authoritative) documents are held to the standard of accuracy.

### Multi-Directory Support

`kizami audit` iterates over all directories configured in `[audit] dirs` (default: both `docs/decisions/` and `docs/design/`):

```
cmd/audit.go:
  dirs := auditDirs(root, cfg)
  for _, dir := range dirs:
    results += Audit(dir, root)
```

The `audit.dirs` config key defaults to `documents.dirs`, ensuring both ADRs and design docs are covered without explicit configuration.

### CI Integration

`kizami init` generates `.github/workflows/adr-audit.yml`, which:
1. Runs `kizami audit` on a weekly schedule (`cron: '0 0 * * 1'`)
2. Also supports manual trigger via `workflow_dispatch`
3. If stale references are found (exit via `stale_found` output), creates a GitHub Issue with the full audit report
4. Deduplicates issues: only one `[kizami audit]` issue is open at a time

### Blame: The Reverse Lookup

`kizami blame <file>` answers the complementary question: "which documents mention this file?"

It runs two passes over the document directory:
1. **Full-text search** (via ripgrep or stdlib fallback): finds documents containing the exact file path string
2. **Directory prefix match** (`blameDirEntries`): finds documents with a directory entry that is a prefix of the queried file path

Results are deduplicated by file path and sorted by document ID.
This is the inverse operation to audit: audit finds documents whose related files are gone; blame finds documents for a file that still exists.

## Open Questions

- **Symbol-level drift**: If a function referenced in a document is renamed but the file still exists, audit does not detect this. This is a known limitation. A future approach could parse the document body for function names and cross-reference them against AST or `ctags` output.
- **Renamed files**: `git mv` renames a file but kizami has no awareness of git history. A future `kizami sync` command could use `git log --follow` to detect renames and suggest updating Related Files automatically.

## Related Files

- `internal/decision/audit.go`
- `internal/search/blame.go`
- `cmd/audit.go`
- `internal/initializer/templates/adr-audit.yml`
- `kizami.toml`
