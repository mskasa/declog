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

This is the documentation problem. **The reasoning behind a decision only exists at the moment the decision is made.** Once the conversation ends, it's gone. No amount of reading the code will bring it back.

kizami is built around this insight. It gives teams a lightweight workflow for capturing the *why* at the right moment, storing it in Git, and keeping it honest as the codebase evolves.

---

## "Just ask AI to update the docs" — is that enough?

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

## AI is a reactive tool

The "just ask AI" approach requires someone to notice that documentation is stale and actually follow through on asking. But by the time someone notices, the original context is already gone.

The reason documentation goes stale in the first place is that no one remembers to do these things. kizami automates the enforcement layer with CI and git hooks — so the prompt happens automatically, not when someone happens to think of it.

AI is a reactive tool: it answers when asked. kizami is proactive infrastructure: it detects change and surfaces it.

---

## In the AI era, accurate documentation is a force multiplier

AI can read code, suggest changes, and write tests. But it cannot know why a design is the way it is, or what trade-offs shaped it. To understand intent, AI needs context — and that context is natural language.

Accurate documentation is a blueprint for AI. The better your docs, the more precisely AI can assist. kizami is the system that keeps those docs accurate.

