# kizami

<p align="center">
  <img src="docs/assets/logo.svg" alt="kizami" width="400">
</p>

**`kizami`** — ドキュメントとコードの乖離を防ぐ、Go製のリビングドキュメント管理CLIツールです。

[English README](README.md)

---

設計上の意思決定は、IssueやPR、Slackに散らばり、やがて失われてしまいがちです。
`kizami` は、その意思決定をMarkdownファイルとしてコードと並べて保存します。すべての判断の理由が、リポジトリの中に永続的に残ります。

```
$ kizami log "SQLiteではなくPostgreSQLを使う"
Created: docs/decisions/0007-sqliteではなくpostgresqlを使う.md

$ kizami list
ID    Date        Status    Title
--    ----        ------    -----
0007  2026-03-12  Active    SQLiteではなくPostgreSQLを使う
0006  2026-03-12  Active    Command Name "kizami"
...

$ kizami search "PostgreSQL"
docs/decisions/0007-sqliteではなくpostgresqlを使う.md:1: # 0007: SQLiteではなくPostgreSQLを使う
```

## インストール

### go install（Goをお持ちの方に推奨）

```bash
go install github.com/mskasa/kizami@latest
```

### バイナリをダウンロード

[Releasesページ](https://github.com/mskasa/kizami/releases)からお使いのプラットフォーム向けの最新バイナリをダウンロードしてください。

**macOS / Linux**

```bash
# macOS (Apple Silicon)
curl -L https://github.com/mskasa/kizami/releases/latest/download/kizami_darwin_arm64.tar.gz | tar xz
mv kizami /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/mskasa/kizami/releases/latest/download/kizami_darwin_amd64.tar.gz | tar xz
mv kizami /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/mskasa/kizami/releases/latest/download/kizami_linux_amd64.tar.gz | tar xz
mv kizami /usr/local/bin/
```

**Windows（PowerShell・管理者権限が必要）**

```powershell
# amd64
Invoke-WebRequest -Uri https://github.com/mskasa/kizami/releases/latest/download/kizami_windows_amd64.zip -OutFile kizami.zip
Expand-Archive kizami.zip -DestinationPath kizami_bin
Move-Item kizami_bin\kizami.exe C:\Windows\System32\kizami.exe
Remove-Item kizami.zip, kizami_bin -Recurse
```

## クイックスタート

```bash
# 1. decisionsディレクトリを初期化する
kizami init

# 2. 意思決定を記録する
kizami log "SQLiteではなくPostgreSQLを使う"
# 生成されたMarkdownファイルが $EDITOR で開きます

# 3. 一覧を表示する
kizami list

# 4. 特定の意思決定を表示する
kizami show 7

# 5. キーワードで検索する
kizami search "PostgreSQL"

# 6. ステータスを更新する
kizami status 7 inactive
kizami status 3 superseded --by 7
```

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `kizami init` | decisionsディレクトリとGitHub Actionsワークフローを初期化する |
| `kizami log "<タイトル>"` | 新しい意思決定記録を作成し、`$EDITOR` で開く |
| `kizami list` | すべての意思決定を新しい順に一覧表示する |
| `kizami show <id>` | 指定した意思決定の全文を表示する |
| `kizami search <キーワード>` | キーワードで意思決定を検索する |
| `kizami status <id> <ステータス>` | 意思決定のステータスを更新する |
| `kizami blame <ファイル>` | 指定ファイルを参照している意思決定を逆引きする |
| `kizami audit` | Related Filesセクションとコードの乖離を検出する |
| `kizami review` | 長期未更新の意思決定を検出する |

### ステータス一覧

| ステータス | 意味 |
|---|---|
| `Active` | 現在有効な意思決定（デフォルト） |
| `Inactive` | 無効になった意思決定（代替なし） |
| `Superseded by NNNN` | 別の意思決定に置き換えられた |

### `kizami status` の使用例

```bash
kizami status 3 inactive
kizami status 3 superseded --by 5   # 0003 が 0005 に置き換えられたことを記録
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
- Status: Active
- Author: あなたの名前

## Context

<!-- なぜこの意思決定が必要だったか -->

## Decision

<!-- 何を決めたか -->

## Consequences

<!-- 影響・メリット・トレードオフ -->

## Alternatives Considered

<!-- 検討したが採用しなかった選択肢とその理由 -->

## Related Files

<!-- この意思決定に関連するファイル（例: internal/db/db.go） -->
```

## 乖離検出

`## Related Files` セクションは、ドキュメントと参照するソースファイルを結ぶ唯一の接点です。
`kizami audit` は、そのファイルが削除・移動されていないかを検証し、ドキュメントの陳腐化を防ぎます。

```bash
kizami audit
# docs/decisions/ 内の全 Related Files エントリをチェックし、存在しないファイルを報告する
```

## 検索について

`kizami search` は、インストールされている場合は [ripgrep](https://github.com/BurntSushi/ripgrep) を優先して使用します。インストールされていない場合はGoの標準ライブラリにフォールバックするため、どの環境でも動作します。

## 設計上の意思決定

このリポジトリは `kizami` 自身を使って設計上の意思決定を記録しています。[`docs/decisions/`](docs/decisions/) を参照してください。

## ライセンス

MIT
