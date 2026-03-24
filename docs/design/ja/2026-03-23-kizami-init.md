# kizami init

- Date: 2026-03-23
- Type: Design
- Status: Active
- Author: masahiro.kasatani

## Overview

`kizami init` は、1つのインタラクティブなコマンドでリポジトリに kizami を導入するためのセットアップを行う。decisions ディレクトリの作成、`kizami.toml` の生成、オプションで CI ワークフローと pre-commit フックのインストールまでをカバーする。

## Background

kizami を新たに導入するユーザーは、ドキュメントディレクトリの作成、設定ファイルの記述、GitHub Actions ワークフローのセットアップ、git フックのインストールなど、複数のセットアップ手順を踏む必要がある。いずれも定型的なボイラープレートを含んでおり、専用コマンドがなければ導入のハードルが高く、設定ミスも起きやすい。

`kizami init` は、各オプションコンポーネントを `y/n` プロンプトで選択しながら進めるインタラクティブなフローに集約し、sensible なデフォルト値と冪等性チェックにより再実行しても安全に使えるようにする。

## Goals / Non-Goals

**Goals:**
- `docs/decisions/` が存在しない場合に作成する
- 全セクションとデフォルト値を含む `kizami.toml` を生成する
- 以下の4つのオプションコンポーネントを `y/n` プロンプトで選択してインストールできる：
  - ADR チェック CI ワークフロー（`adr-check.yml`）
  - pre-commit フック
  - 週次 audit CI ワークフロー（`adr-audit.yml`）
  - 自動プロモートワークフロー（`kizami-promote.yml`）
- 冪等性：既に存在するファイルはスキップ（警告を表示）
- グローバル設定は `kizami init --global` で `~/.config/kizami/config.toml` を生成

**Non-Goals:**
- 非インタラクティブ/サイレントモード（現時点では全プロンプトが必須）
- 初回 ADR の自動生成
- 既存の config やワークフローの更新管理

## Design

### 初期化フロー

`internal/initializer/init.go` の `Initializer.Run()` が以下のステップを順番に実行する：

```
kizami init
  │
  ├── 1. createDecisionsDir()     — docs/decisions/ を作成（存在する場合はスキップ）
  ├── 2. setupWorkflow()          — プロンプト: adr-check.yml
  ├── 3. setupHook()              — プロンプト: pre-commit フック
  ├── 4. setupAuditWorkflow()     — プロンプト: adr-audit.yml
  ├── 5. setupPromoteWorkflow()   — プロンプト: kizami-promote.yml
  └── 6. setupConfig()            — kizami.toml を書き込む（存在する場合はスキップ）
```

各ステップは独立しており、あるステップが失敗しても後続のステップはスキップされない（そのステップ自体はエラーを返す）。`y/n` プロンプトは `os.Stdin` 上の単一の `bufio.Scanner` を共有し、入力ストリームを正しく消費する。

### 冪等性

各ステップはファイル書き込み前に対象パスに対して `os.Stat` を呼び出す：
- ファイルが既に存在する → `⚠️` 警告を表示してスキップ
- 存在しない → 作成して `✅` 確認メッセージを表示

これにより、部分的なセットアップ後（例：最初に audit ワークフローを `n` にしたが後から追加したい場合）でも `kizami init` を安全に再実行できる。

### 生成される成果物

#### `kizami.toml`

デフォルト設定ファイルは `init.go` 内に Go の文字列定数としてハードコードされている。全設定セクションとデフォルト値を含む：

```toml
[ai]
model = "claude-sonnet-4-20250514"

[documents]
dirs = ["docs/decisions", "docs/design"]

[decisions]
dir = "docs/decisions"

[design]
dir = "docs/design"

[audit]
dirs = ["docs/decisions", "docs/design"]

[review]
months_threshold = 6

[editor]
command = "code --wait"
```

#### `.github/workflows/adr-check.yml`

プルリクエストごとに実行される。ソースファイル（非 docs・非 config パス）に触れるコミットに、対応するドキュメント変更が含まれているかどうかをチェックする。デフォルトではマージをブロックせず、リマインダーとして機能する設計。

#### `.git/hooks/pre-commit`

`//go:embed templates/pre-commit` で埋め込まれたシェルスクリプト。コミット前に kizami の利用可能性をチェックし、意思決定記録の作成を検討するよう開発者に促す。既に pre-commit フックが存在する場合は、スクリプト内容を stdout に出力して手動での追記を促す（既存フックを上書きすると他のツールのフックを無言で壊してしまうため）。

#### `.github/workflows/adr-audit.yml`

週次スケジュール（`cron: '0 0 * * 1'`）と `workflow_dispatch` で `kizami audit` を実行する。陳腐化した参照が見つかった場合、`[kizami audit]` タグ付きの GitHub Issue を作成する（重複防止のため、同タグの Issue は1つのみオープン状態に保つ）。詳細は Audit and Drift Detection 設計書を参照。

#### `.github/workflows/kizami-promote.yml`

`main` へのプッシュ時に実行される。`Status: Draft` のドキュメントを自動的に `Status: Active` に昇格させる。プロモーションロジックを説明するインラインコメントが含まれており、チームがトリガーをカスタマイズしたり無効化したりできるようになっている。

### テンプレートの埋め込み

すべてのワークフローとフックのテンプレートは、Go の `//go:embed` ディレクティブによりビルド時に埋め込まれる：

```go
//go:embed templates/adr-check.yml
var adrCheckWorkflow string

//go:embed templates/adr-audit.yml
var adrAuditWorkflow string

//go:embed templates/kizami-promote.yml
var promoteWorkflow string

//go:embed templates/pre-commit
var hookScript string
```

バイナリに埋め込まれるため、実行時に外部テンプレートファイルは不要。テンプレートを更新する場合はバイナリの再ビルドが必要。

### `Initializer` 構造体

```go
type Initializer struct {
    Root   string
    Input  io.Reader
    Output io.Writer
}
```

`Input` と `Output` を `os.Stdin` / `os.Stdout` にハードコードせず注入することで、実際のターミナルなしに完全にテスト可能な設計になっている。

## Open Questions

- **非インタラクティブモード**: `--yes` フラグで全プロンプトを自動承認する機能は、CI ベースのセットアップスクリプトで有用だが、未実装。
- **Config の更新**: `kizami.toml` が既に存在する場合、後続バージョンで追加された新しいキーがあっても `kizami init` は完全にスキップする。将来的な `kizami init --upgrade` で既存設定に新しいデフォルト値をマージする機能が考えられる。
- **`docs/design/` の作成**: `createDecisionsDir` は `docs/decisions/` のみを作成する。`docs/design/` ディレクトリは `kizami init` では作成されない（既知の欠落）。

## Related Files

- `internal/initializer/init.go`
- `internal/initializer/hook.go`
- `internal/initializer/templates/adr-check.yml`
- `internal/initializer/templates/adr-audit.yml`
- `internal/initializer/templates/kizami-promote.yml`
- `internal/initializer/templates/pre-commit`
- `cmd/init.go`
