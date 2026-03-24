---
layout: default
title: 開発ワークフロー
parent: 日本語
nav_order: 1
---

# 開発ワークフロー

kizamiを日常の開発プロセスにどう組み込むかを説明します。

[← ドキュメントトップへ](.)

---

## 基本の流れ

意味のある技術的判断を伴うコード変更を行ったときは、**コード変更と同じコミット**にドキュメントを含めましょう。

{% raw %}
<div class="mermaid">
flowchart TD
    A[コードを変更] --> B[git add]
    B --> C["kizami adr / kizami design"]
    C --> D{"類似ADRはある?"}
    D -- ある --> E[既存ADRを確認し必要なら Superseded に]
    D -- ない --> F[テンプレートを記入]
    E --> F
    F --> G[新しいドキュメントをgit add]
    G --> H[コードとドキュメントをまとめてgit commit]
    H --> I["kizami blame / search で過去の判断を参照"]
</div>
{% endraw %}

---

## ステップごとの解説

### 1. コードを変更する

通常通りコードを変更します。準備ができたら `git add` でステージングします。

```bash
git add internal/db/db.go
```

### 2. ADRを作成する

`kizami adr` に意思決定を表すタイトルをつけて実行します。

```bash
kizami adr "データベースアクセスにコネクションプールを使う"
```

kizamiは以下を自動で行います：
- **ステージ済みファイルを自動挿入** → `## Related Files` セクションへ
- **類似ADRを表示** → タイトルが部分一致する既存ドキュメントがあれば提示
- 生成されたファイルを `$EDITOR` で開く

`--ai` フラグを付けると、ステージ済みの差分をもとにAIがドラフトを自動生成します：

```bash
kizami adr --ai "データベースアクセスにコネクションプールを使う"
```

### 3. 既存の意思決定を更新する（必要な場合）

新しいADRが既存のADRを置き換える場合は、コミット前に Superseded 状態にしておきます。

```bash
kizami supersede 2026-03-01-use-single-db-connection --by 2026-03-12-use-connection-pooling
```

### 4. コードとドキュメントをまとめてコミット

コード変更と新しいADRを同じコミットに含めます。こうすることで、意思決定と実装がGitの履歴上で常に紐付いた状態になります。

```bash
git add docs/decisions/2026-03-12-use-connection-pooling.md
git commit -m "feat: データベースアクセスにコネクションプールを追加"
```

### 5. 過去の判断を参照する

いつでも `kizami blame` や `kizami search` を使って、なぜそのように実装されたかを追跡できます。

```bash
# 特定のファイルを参照しているADRを逆引き
kizami blame internal/db/db.go

# キーワードで検索
kizami search "コネクションプール"
```

---

## 定期メンテナンス

### 陳腐化したドキュメントを検出する

```bash
kizami review
```

長期間更新されていないADRを一覧表示します（閾値は設定で変更可能）。定期的なチームレビューに活用できます。

### ドキュメントとコードの乖離を検出する

```bash
kizami audit
```

全ドキュメントの `## Related Files` エントリを確認します。参照されているファイルが削除・移動されていた場合に報告します。

`kizami init` で生成されるGitHub Actionsワークフローで自動化することもできます：

```bash
kizami init
# → .github/workflows/kizami-audit.yml が生成される
```

---

## 設計ドキュメントとADRの使い分け

| | ADR（`kizami adr`） | 設計ドキュメント（`kizami design`） |
|---|---|---|
| **目的** | 意思決定の*理由*を記録 | 設計の*内容*を記述 |
| **デフォルトステータス** | Draft | Draft |
| **デフォルトディレクトリ** | `docs/decisions/` | `docs/design/` |
| **典型的なライフサイクル** | Draft → Active → （Inactive / Superseded） | Draft → Active |
| **auditの対象** | あり（Active のみ） | あり（Active のみ） |

どちらも同じMADR互換テンプレートを使用し、`## Related Files` をサポートしています。
