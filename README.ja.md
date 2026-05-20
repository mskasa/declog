# kizami

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./docs/assets/kizami-logo-dark.svg">
    <img src="./docs/assets/kizami-logo-light.svg" alt="kizami logo" width="400">
  </picture>
</p>

**`kizami`** — ADRや設計ドキュメントをコードと並べて管理し、乖離を自動検出するGo製のCLIツールです。

[English README](README.md)

---

設計上の意思決定は、IssueやPR、Slackに散らばり、やがて失われてしまいがちです。
`kizami` は、その意思決定をMarkdownファイルとしてコードと並べて保存します。すべての判断の理由が、コードの隣に記録され、正確に維持されます。

```
$ kizami adr "use PostgreSQL over SQLite"
Created: docs/decisions/2026-03-12-use-postgresql-over-sqlite.md

$ kizami list
Slug                            Date        Status    Title
----                            ----        ------    -----
use-postgresql-over-sqlite      2026-03-12  Draft     use PostgreSQL over SQLite
command-name-kizami             2026-03-12  Active    Command name "kizami"
...

$ kizami search "PostgreSQL"
docs/decisions/2026-03-12-use-postgresql-over-sqlite.md:1: # use PostgreSQL over SQLite
```

## ドキュメント

詳細なドキュメントは **[mskasa.github.io/kizami](https://mskasa.github.io/kizami/)** で参照できます。ワークフローガイド・ADRの書き方・設定リファレンスなどを掲載しています。

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

# 2. ADRを作成する
kizami adr "SQLiteではなくPostgreSQLを使う"
# 生成されたMarkdownファイルが $EDITOR で開きます

# 2b. 設計ドキュメントを作成する
kizami design "コネクションプール設計"

# 3. 一覧を表示する
kizami list

# 4. 特定の意思決定を表示する
kizami show use-postgresql-over-sqlite

# 5. キーワードで検索する
kizami search "PostgreSQL"

# 6. ステータスを更新する
kizami status use-postgresql-over-sqlite inactive
kizami status use-sqlite superseded --by use-postgresql-over-sqlite
```

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `kizami init` | decisionsディレクトリとGitHub Actionsワークフローを初期化する |
| `kizami adr "<タイトル>"` | 新しいADRを作成し、`$EDITOR` で開く |
| `kizami design "<タイトル>"` | 新しい設計ドキュメントを作成し、`$EDITOR` で開く |
| `kizami list` | すべてのドキュメントを新しい順に一覧表示する |
| `kizami show <slug>` | 指定したドキュメントの全文を表示する |
| `kizami search <キーワード>` | キーワードでドキュメントを検索する |
| `kizami status <slug> <ステータス>` | ドキュメントのステータスを更新する |
| `kizami supersede <slug> "<タイトル>"` | 既存のドキュメントを置き換え、新しいドキュメントを作成する |
| `kizami blame <ファイル>` | 指定ファイルを参照しているドキュメントを逆引きする |
| `kizami audit` | Related Filesセクションとコードの乖離を検出する |
| `kizami lint` | CIでドキュメント構造を検証する |
| `kizami review` | 長期未更新のドキュメントを検出する |

### ステータス一覧

| ステータス | 意味 |
|---|---|
| `Active` | 現在有効な意思決定（デフォルト） |
| `Inactive` | 無効になった意思決定（代替なし） |
| `Superseded by <slug>` | 別のドキュメントに置き換えられた |

### `kizami status` の使用例

```bash
kizami status use-sqlite inactive
kizami status use-sqlite superseded --by use-postgresql-over-sqlite
```

## 意思決定ファイルのフォーマット

意思決定は [MADR](https://adr.github.io/madr/) 互換のテンプレートを使い、`docs/decisions/` 以下にMarkdownファイルとして保存されます。

```
docs/decisions/
├── 2026-03-12-use-go-over-shell-script.md
├── 2026-03-12-use-cobra-for-cli-framework.md
└── ...
```

ファイル名は `YYYY-MM-DD-kebab-case-title.md` の形式です。

```markdown
# use PostgreSQL over SQLite

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
