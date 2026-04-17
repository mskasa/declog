---
layout: default
title: ADR Guide
nav_order: 4
---

# ADR Guide

This page explains how to write effective ADRs with kizami — what to capture, how to fill in the template, and how to manage statuses over time.

[← Back to Documentation](.)

---

## What is an ADR?

An **Architecture Decision Record (ADR)** is a short document that captures a significant technical decision: what was decided, why, and what the consequences are.

The key insight is simple: **the *why* behind a decision is just as important as the decision itself** — yet it's the first thing to be lost. ADRs keep that reasoning alive, in the repository, next to the code.

kizami uses a [MADR](https://adr.github.io/madr/)-compatible template, which keeps ADRs concise and consistent.

---

## What deserves an ADR?

### Record these decisions

- Technology selection (libraries, frameworks, databases, file formats)
- Choices between multiple implementation approaches
- Decisions driven by external constraints (performance tests, incidents, security requirements)
- Changes that retire or replace an existing design
- Decisions that affect multiple files or components

### Skip the ADR for these

- Variable or function naming
- Self-evident implementation details
- Reasoning that fits entirely within a single file (use a code comment instead)

### ADR vs. code comment

| Scope | Where to document |
|---|---|
| Reasoning contained within a single file | Code comment |
| Reasoning that spans multiple files or components | ADR |
| Both | Write both; leave a link to the ADR in the comment |

```go
// AuthorFromGit reads the author name from git config.
// See docs/decisions/2026-03-16-use-git-config-for-author.md for the rationale.
func AuthorFromGit() string { ... }
```

---

## The template

```markdown
# Title

- Date: YYYY-MM-DD
- Status: Draft
- Author: your name

## Context

Why this decision was needed. Describe the background, constraints, and problem.

## Decision

What was decided. State clearly in 1–3 sentences.

## Consequences

Impact, benefits, and trade-offs of this decision.

## Alternatives Considered

Options that were considered but not adopted, and why. (Optional)

## Related Files

List files related to this decision (e.g. internal/search/search.go).
```

### AI-assisted drafting

Add the `--ai` flag to have AI generate a draft for you. kizami reads the diff of staged files and pre-fills the Context, Decision, and Consequences sections before opening your editor.

```bash
kizami adr --ai "use connection pooling for database access"
kizami design --ai "connection pool design"
```

Treat the generated content as a starting point — always review and edit before committing.

### Writing tips

**Context** — Focus on the *problem*, not the solution. Why was a decision needed at all? What constraints existed?

**Decision** — Be direct. "We will use X" rather than "We considered using X". One to three sentences is ideal.

**Consequences** — Be honest about trade-offs. A good ADR acknowledges what you're giving up, not just what you're gaining.

**Related Files** — List the source files most closely tied to this decision. This powers `kizami blame` and `kizami audit`. You can list directories too — kizami will treat all files under them as related.

---

## Status management

### Status values

| Status | Meaning | When to use |
|---|---|---|
| `Draft` | Being written or not yet implemented | Default on creation |
| `Active` | Currently valid decision | After the change is implemented and merged |
| `Inactive` | No longer applicable | When a decision becomes invalid with no replacement |
| `Superseded by <slug>` | Replaced by another ADR | When a new ADR takes over |

### Typical lifecycle

```
Draft → Active → Inactive
                ↘ Superseded by YYYY-MM-DD-new-decision
```

### Updating status

```bash
kizami status 2026-03-12-use-sqlite active
kizami status 2026-03-12-use-sqlite inactive
kizami status 2026-03-12-use-sqlite superseded --by 2026-06-01-use-postgresql
```

### Automating Draft → Active promotion

`kizami init` generates a workflow (`kizami-promote.yml`) that automatically promotes `Draft` documents to `Active` on push to main. This lets you commit ADRs as `Draft` during development and have them automatically become `Active` when the change lands.

```bash
kizami init
# → .github/workflows/kizami-promote.yml (generated but commented out)
# Edit the file to enable the workflow
```

---

## Updating ADRs

ADRs can be updated directly. Git manages the history — `git diff` shows what changed, `git log` shows why.

**Permitted changes:**
- Revising the content when the same decision is refined (direct update is fine)
- Updating status
- Fixing typos
- Adding entries to the Related Files section

**When to create a new ADR instead:**
When the *direction* of a decision changes entirely, create a new ADR and mark the old one as `Superseded by <slug>`.

**Good commit messages when updating:**
```
docs: update ADR use-postgresql — increase pool size from 10 to 20 based on load test
```

---

## File naming

kizami generates filenames automatically:

```
YYYY-MM-DD-kebab-case-title.md
```

Examples:
```
2026-03-12-use-go-over-shell-script.md
2026-06-01-switch-to-postgresql.md
```

The date prefix ensures chronological sort order. The title is automatically converted to lowercase kebab-case.

### Bringing in existing documents

If your team already has Markdown documents with established names (e.g. `ARCHITECTURE.md`, `API-SPEC.md`), you do not need to rename them. kizami recognises any `.md` file as a managed document as long as it contains **both**:

- A line beginning with `- Status:`
- A `## Related Files` section

Add these two markers to an existing file, and it becomes visible to `kizami list`, `kizami audit`, and all other commands. Its slug is the filename without the `.md` extension (e.g. `ARCHITECTURE`).

### Managing non-Markdown files

For files that cannot carry kizami markers — CSV, YAML, SQL, images, etc. — place a `.kizami` sidecar file alongside them:

```yaml
# data/test_matrix.csv.kizami
title: Test matrix for user flow
date: 2026-04-17
author: your name
related:
  - tests/user_flow_test.go
```

The sidecar file is fully supported by `kizami list`, `kizami show`, `kizami blame`, and `kizami audit`. Its slug is the managed filename (e.g. `test_matrix.csv`).

Sidecars have no `status` field — they are always included in `kizami audit`. Each managed file gets its own sidecar — one file, one sidecar.
