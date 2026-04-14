---
layout: default
title: ADR運用ガイド
parent: 日本語
nav_order: 2
---

# ADR運用ガイド

kizamiを使った効果的なADRの書き方を説明します。何を記録するか、テンプレートの各セクションの書き方、そしてステータスの運用方法について解説します。

[← ドキュメントトップへ](.)

---

## ADRとは？

**ADR（Architecture Decision Record）** は、重要な技術的意思決定を記録した短いドキュメントです。何を決めたか、なぜそう決めたか、そしてその結果どうなったかを記録します。

重要な考え方：**意思決定の*理由*は、決定そのものと同じくらい重要です** — にもかかわらず、最初に失われるのが理由です。ADRはその理由をコードの隣に、リポジトリの中に生き続けさせます。

kizamiは [MADR](https://adr.github.io/madr/) 互換のテンプレートを使用しており、ADRを簡潔かつ一貫した形式に保ちます。

---

## ADRの対象となる判断・ならない判断

### ADRに記録すべき判断

- 技術選定（ライブラリ・フレームワーク・データベース・ファイル形式）
- 複数の実装アプローチからの選択
- 外部要因による制約から生まれた決定（パフォーマンステスト・インシデント・セキュリティ要件）
- 既存設計を廃止・置き換える変更
- 複数のファイルやコンポーネントにまたがる決定

### ADRを書かなくてよい判断

- 変数名・関数名の命名
- 自明な実装の詳細
- 1つのファイル内に収まる理由（コードコメントで十分）

### ADRとコードコメントの使い分け

| スコープ | どこに書くか |
|---|---|
| 1つのファイル内に収まる理由 | コードコメント |
| 複数のファイル・コンポーネントにまたがる理由 | ADR |
| 両方にまたがる | 両方に書く。コードコメントにADRへのリンクを残す |

```go
// AuthorFromGit は git config からauthor名を読み取ります。
// 環境変数の代わりにgit configを使う理由: docs/decisions/2026-03-16-use-git-config-for-author.md
func AuthorFromGit() string { ... }
```

---

## テンプレート

```markdown
# タイトル

- Date: YYYY-MM-DD
- Status: Draft
- Author: あなたの名前

## Context

なぜこの意思決定が必要だったか。背景・制約・問題を記述する。

## Decision

何を決めたか。1〜3文で明確に述べる。

## Consequences

この決定の影響・メリット・トレードオフ。

## Alternatives Considered

検討したが採用しなかった選択肢とその理由。（省略可）

## Related Files

この意思決定に関連するファイル（例: internal/search/search.go）
```

### AIによるドラフト生成

`--ai` フラグを付けると、AIがテンプレートの下書きを自動生成します。ステージ済みファイルの差分を読み込み、Context・Decision・Consequencesのセクションを埋めた状態でエディタが開きます。

```bash
kizami adr --ai "データベースアクセスにコネクションプールを使う"
kizami design --ai "コネクションプール設計"
```

生成された内容はあくまで出発点です。コミット前に必ず内容を確認・編集してください。

### 書き方のポイント

**Context** — *問題*にフォーカスする。解決策ではなく、なぜ決断が必要だったかを書く。どんな制約があったか？

**Decision** — 「〜にする」と直接的に書く。「〜を検討した」ではなく。1〜3文が理想的。

**Consequences** — トレードオフに正直に。良いADRは得るものだけでなく、諦めることも認める。

**Related Files** — この意思決定と最も深く関わるソースファイルを列挙する。`kizami blame` と `kizami audit` の精度に直結する。ディレクトリを指定することもでき、その場合は配下の全ファイルが対象になる。

---

## ステータスの運用

### ステータス一覧

| ステータス | 意味 | 使うタイミング |
|---|---|---|
| `Draft` | 作成中・未実装 | 作成直後のデフォルト |
| `Active` | 現在有効な意思決定 | 変更が実装されマージされた後 |
| `Inactive` | 無効になった（代替なし） | 代替のADRなく無効化された時 |
| `Superseded by <slug>` | 別のADRに置き換えられた | 新しいADRが引き継ぐ時 |

### 典型的なライフサイクル

```
Draft → Active → Inactive
                ↘ Superseded by YYYY-MM-DD-new-decision
```

### ステータスの更新

```bash
kizami status 2026-03-12-use-sqlite active
kizami status 2026-03-12-use-sqlite inactive
kizami status 2026-03-12-use-sqlite superseded --by 2026-06-01-use-postgresql
```

### Draft → Active の自動昇格

`kizami init` で生成されるワークフロー（`kizami-promote.yml`）を有効にすると、main ブランチへの push 時に `Draft` ステータスのドキュメントを自動的に `Active` に昇格させることができます。

```bash
kizami init
# → .github/workflows/kizami-promote.yml が生成される（コメントアウト済み）
# ファイルを編集してワークフローを有効化する
```

コミット時点で Draft のまま残し、マージされた時点で Active になるフローを自動化したい場合に活用してください。

---

## ADRの更新原則

ADRは直接更新して構いません。Gitが履歴を管理します — `git diff` で何が変わったかがわかり、`git log` でなぜ変わったかがわかります。

**更新して良いもの：**
- 同じ決定を洗練・修正する場合（直接更新で十分）
- ステータスの更新
- 誤字修正
- Related Filesセクションへの追記

**新しいADRを作成する場合：**
決定の*方向性*が根本的に変わる場合は、新しいADRを作成し、古いADRを `Superseded by <slug>` にマークします。

**更新時のコミットメッセージ：**
```
docs: ADR use-postgresql を更新 — 負荷テストの結果、プールサイズを10→20に変更
```

---

## ファイル命名規則

kizamiはファイル名を自動で生成します：

```
YYYY-MM-DD-kebab-case-title.md
```

例：
```
2026-03-12-use-go-over-shell-script.md
2026-06-01-switch-to-postgresql.md
```

日付プレフィックスにより時系列でソートされます。タイトルは自動的に小文字のkebab-caseに変換されます。

### 既存ドキュメントの取り込み

チームにすでに確立されたファイル名のMarkdownドキュメント（`ARCHITECTURE.md`、`API-SPEC.md` など）がある場合、リネームは不要です。kizamiは以下の **両方** を含む `.md` ファイルを管理対象ドキュメントとして認識します：

- フロントマターに `- Status:` 行がある
- `## Related Files` セクションがある

この2つのマーカーを既存ファイルに追加するだけで、`kizami list`・`kizami audit` などすべてのコマンドから参照できるようになります。slugは `.md` 拡張子を除いたファイル名になります（例：`ARCHITECTURE`）。
