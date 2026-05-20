---
layout: default
title: Why kizami?
nav_order: 2
---

# Why kizami?

[日本語版はこちら](ja/why-kizami)

---

## Why documentation gets lost

You've been here before. You open a file written six months ago and stare at an unusual pattern. Why is it this way?

`git log` says "fix: update logic." The PR has no description. Slack: the message is gone. The person who wrote it has since left the team.

The code works fine. But nobody knows *why* it works that way. Changing it feels risky. Code review takes longer than it should. Onboarding new teammates means explaining things that should already be written down.

This is the documentation problem — and it's not a matter of effort or discipline. It's a matter of timing. **The reasoning behind a decision only exists at the moment the decision is made.** Once the conversation ends, it's gone. No amount of reading the code will bring it back.

kizami is built around this insight. It gives teams a lightweight workflow for capturing the *why* at the right moment, storing it in Git, and keeping it honest as the codebase evolves.

---

## "Just ask AI to read the code and update the docs" — is that enough?

With AI assistants able to read entire codebases, a reasonable question arises:
**why bother with a dedicated tool when you can just ask AI to keep your docs up to date?**

The short answer: AI can tell you *what* the code does today. It cannot tell you *why* a decision was made.

---

## AI reads the result. kizami captures the reasoning.

Code only preserves the outcome of a decision — not the thinking behind it.

Things AI cannot recover from reading code:

- A choice driven by load test results from three months ago
- Why you chose SQLite over PostgreSQL — the code only shows SQLite
- The alternatives your team discussed and rejected
- The constraints that existed at the time: budget, deadline, team skill set

**The reasoning behind a decision can only be written at the moment the decision is made.** kizami captures the *why* at the right time, stores it in Git, and keeps it honest over time.

---

## "Just ask AI" assumes someone remembers to ask

For the "ask AI" approach to work, three things need to happen:

1. Someone notices the documentation is stale
2. Someone remembers to ask AI about it
3. Someone actually follows through

The reason documentation goes stale in the first place is that no one remembers to do these things. kizami automates the enforcement layer with CI and git hooks — so the prompt happens automatically, not when someone happens to think of it.

AI is a reactive tool: it answers when asked. kizami is proactive infrastructure: it detects change and surfaces it.

---

## kizami makes AI output permanent

When you use `kizami adr --ai`, AI reads the staged diff and drafts the Context, Decision, and Consequences sections for you. The difference from a chat session: the output is saved to a file, committed to Git, and becomes part of the repository's auditable history.

kizami is not a replacement for AI — it's the layer that makes AI-generated documentation **persistent, versioned, and verifiable**.

