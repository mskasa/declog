# declog

**`why`** — アーキテクチャ上の意思決定を記録・検索するための最小限のCLIツールです。

[English README](README.md)

---

設計上の意思決定は、IssueやPR、Slackに散らばり、やがて失われてしまいがちです。
`why` は、その意思決定をMarkdownファイルとしてコードと並べて保存します。すべての判断の理由が、リポジトリの中に永続的に残ります。

```
$ why log "SQLiteではなくPostgreSQLを使う"
Created: docs/decisions/0007-sqliteではなくpostgresqlを使う.md

$ why list
ID    Date        Status    Title
--    ----        ------    -----
0007  2026-03-12  Proposed  SQLiteではなくPostgreSQLを使う
0006  2026-03-12  Accepted  Command Name "why"
...

$ why search "PostgreSQL"
docs/decisions/0007-sqliteではなくpostgresqlを使う.md:1: # 0007: SQLiteではなくPostgreSQLを使う
```

## インストール

### バイナリをダウンロード（推奨）

[Releasesページ](https://github.com/mskasa/declog/releases)からお使いのプラットフォーム向けの最新バイナリをダウンロードしてください。

```bash
# macOS (Apple Silicon)
curl -L https://github.com/mskasa/declog/releases/latest/download/why_darwin_arm64.tar.gz | tar xz
mv why /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/mskasa/declog/releases/latest/download/why_darwin_amd64.tar.gz | tar xz
mv why /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/mskasa/declog/releases/latest/download/why_linux_amd64.tar.gz | tar xz
mv why /usr/local/bin/
```

### go install

```bash
go install github.com/mskasa/declog@latest
```

## クイックスタート

```bash
# 1. 意思決定を記録する
why log "SQLiteではなくPostgreSQLを使う"
# 生成されたMarkdownファイルが $EDITOR で開きます

# 2. 一覧を表示する
why list

# 3. 特定の意思決定を表示する
why show 7

# 4. キーワードで検索する
why search "PostgreSQL"

# 5. ステータスを更新する
why status 7 accepted
why status 3 superseded --by 7
```

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `why log "<タイトル>"` | 新しい意思決定記録を作成し、`$EDITOR` で開く |
| `why list` | すべての意思決定を新しい順に一覧表示する |
| `why show <id>` | 指定した意思決定の全文を表示する |
| `why search <キーワード>` | キーワードで意思決定を検索する |
| `why status <id> <ステータス>` | 意思決定のステータスを更新する |

### ステータス一覧

| ステータス | 意味 |
|---|---|
| `Proposed` | 検討中（デフォルト） |
| `Accepted` | 承認・採用済み |
| `Superseded` | 別の意思決定に置き換えられた |
| `Deprecated` | 廃止済み |

### `why status` の使用例

```bash
why status 3 accepted
why status 3 superseded --by 5   # 0003 が 0005 に置き換えられたことを記録
```

## 意思決定ファイルのフォーマット

意思決定は [MADR](https://adr.github.io/madr/) 互換のテンプレートを使い、`docs/decisions/` 以下にMarkdownファイルとして保存されます。

```
docs/decisions/
├── 0001-use-go-over-shell-script.md
├── 0002-use-cobra-for-cli-framework.md
└── ...
```

ファイル名は `NNNN-kebab-case-title.md` の形式で、番号は自動でインクリメントされます。

```markdown
# 0007: SQLiteではなくPostgreSQLを使う

- Date: 2026-03-12
- Status: Proposed
- Author: あなたの名前

## Context

<!-- なぜこの意思決定が必要だったか -->

## Decision

<!-- 何を決めたか -->

## Consequences

<!-- 影響・メリット・トレードオフ -->

## Alternatives Considered

<!-- 検討したが採用しなかった選択肢とその理由 -->
```

## 検索について

`why search` は、インストールされている場合は [ripgrep](https://github.com/BurntSushi/ripgrep) を優先して使用します。インストールされていない場合はGoの標準ライブラリにフォールバックするため、どの環境でも動作します。

## 設計上の意思決定

このリポジトリは `why` 自身を使って設計上の意思決定を記録しています。[`docs/decisions/`](docs/decisions/) を参照してください。

## ライセンス

MIT
