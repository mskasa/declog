---
layout: default
title: 日本語
nav_order: 6
has_children: true
---

<p align="center">
  <img src="{{ site.baseurl }}/assets/kizami-logo-dark.svg" alt="kizami logo" width="480">
</p>

> ドキュメントに、嘘をつかせるな。

**kizami** は、コードと並べてリビングドキュメントを管理し、乖離を自動検出するミニマルなCLIツールです。

設計上の意思決定は、IssueやPR、Slackに散らばり、やがて失われてしまいがちです。
kizamiは、その意思決定をMarkdownファイルとしてコードと並べて保存します。すべての判断の理由が、リポジトリの中に永続的に残ります。

[English version](../)

---

## はじめての方へ

インストール方法やクイックスタートをお探しの方は：

→ [GitHubのREADME](https://github.com/mskasa/kizami#readme) をご覧ください

---

## ドキュメント

このサイトは、kizamiを導入済みの方向けです。日常の開発での使い方や、チームでの運用方法を説明します。

| ページ | 内容 |
|---|---|
| [なぜkizamiが必要か？](why-kizami) | AIアシスタントがある時代になぜkizamiが必要なのか |
| [開発ワークフロー](workflow) | kizamiを日常の開発プロセスに組み込む方法 |
| [ADR運用ガイド](adr-guide) | ADRの書き方・粒度・ステータス管理 |
| [ベストプラクティス](best-practices) | kizamiを最大限に活用するためのヒント |

---

## kizamiとは？

kizamiは2種類のドキュメントを管理します。

**ADR（Architecture Decision Record）** — 技術的な意思決定の*理由*を記録します。
デフォルトで `docs/decisions/` 以下に保存されます。

**設計ドキュメント** — *どのように*設計するかを記録します。
デフォルトで `docs/design/` 以下に保存されます。

どちらも `## Related Files` セクションでソースファイルと紐付けることができます。
`kizami audit` は、参照されたファイルが削除・移動されていないかを検出し、ドキュメントの陳腐化を防ぎます。

```bash
$ kizami adr "SQLiteではなくPostgreSQLを使う"
Created: docs/decisions/2026-03-12-use-postgresql-over-sqlite.md

$ kizami audit
✓ All related files exist.
```
