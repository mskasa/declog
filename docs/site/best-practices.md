---
layout: default
title: Best Practices
nav_order: 5
---

# Best Practices

Practical tips for getting the most out of kizami.

[← Back to Documentation](.)

---

## Commit code and document together

The most important habit: **include the ADR in the same commit as the code change it documents.**

```bash
# Good — decision and implementation are linked in git history
git add internal/db/db.go docs/decisions/2026-03-12-use-connection-pooling.md
git commit -m "feat: add connection pooling for database access"

# Avoid — the decision gets separated from the implementation
git commit -m "feat: add connection pooling"
git commit -m "docs: add ADR for connection pooling"
```

When they're in the same commit, `git log --follow` on any source file will surface the ADR alongside the change. This makes archaeological work much easier months later.

---

## Fill in Related Files carefully

The `## Related Files` section is what makes `kizami blame` and `kizami audit` useful. The more accurately you fill it in, the more value you get.

```markdown
## Related Files

internal/db/db.go
internal/db/pool.go
internal/config/config.go
```

kizami auto-inserts staged files when you run `kizami adr`, but always review and complete the list before committing. Include files that are *conceptually* related, not just the ones you happened to stage.

You can also list directories — kizami will treat all files under them as related:

```markdown
## Related Files

internal/db/
```

---

## Use `kizami review` periodically

Run `kizami review` as part of your team's regular retrospective or sprint planning. It surfaces ADRs that haven't been updated in a long time — a useful prompt to ask whether they're still accurate.

```bash
kizami review
# Lists documents not updated in the last 180 days (configurable)
```

This is especially valuable for ADRs about third-party services, infrastructure, or technology choices that may have evolved.

---

## Automate drift detection with CI

`kizami audit` is most powerful when it runs automatically. Use `kizami init` to generate a GitHub Actions workflow that runs `kizami audit` on a schedule:

```bash
kizami init
# → .github/workflows/kizami-audit.yml
```

The generated workflow runs weekly and opens a GitHub Issue automatically if drift is detected. This ensures nothing slips through the cracks even if nobody remembers to run it manually.

---

## Keep ADRs short

An ADR that nobody reads is worse than no ADR at all. Aim for documents that can be read in two minutes.

- **Context**: 2–4 sentences explaining the problem
- **Decision**: 1–3 sentences stating what was decided
- **Consequences**: A short bulleted list of trade-offs

The goal is to capture the *key insight* — the thing that would take 30 minutes to reconstruct from code alone — not to write a comprehensive design doc. For deeper design exploration, use `kizami design` instead.

---

## Write the ADR before you're sure

The best time to write an ADR is *during* the decision process, not after. If you're considering two approaches, sketch out the ADR before choosing — it forces you to articulate the trade-offs clearly, and often makes the decision itself easier.

Use `Status: Draft` for decisions still in progress.

---

## Use `kizami blame` during code review

When reviewing a PR that modifies a file, check whether any ADRs reference that file:

```bash
kizami blame internal/auth/middleware.go
```

This surfaces the *history* of decisions that shaped the current design. It's an effective way to catch changes that violate the intent of a previous decision — before they land in main.

---

## Frequently Asked Questions

**Q: How many ADRs is too many?**

There's no fixed limit, but if you're creating more than one or two ADRs per week, consider whether some decisions are too small to warrant one. A good heuristic: if a future developer reading the code could figure out *why* from context alone, skip the ADR.

**Q: Should I write ADRs for decisions made in the past?**

Yes, ADRs written after the fact are valuable — especially for decisions that are still shaping the codebase today. Use the actual decision date (even if it was years ago) in the `Date` field, and note in the Context section that this was recorded after the decision was made.

**Q: Can I use kizami for non-code decisions?**

Yes. ADRs work for any significant decision: team processes, API contracts, deployment strategies, data retention policies. As long as the decision is worth preserving and has related files (or can be left without them), kizami can track it.

**Q: What if I disagree with an old ADR?**

Don't delete it — create a new ADR explaining the new direction and mark the old one as `Superseded by <slug>`. The history of the disagreement and resolution is itself valuable context.
