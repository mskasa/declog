# Git フックと CI 連携

- Date: 2026-03-23
- Type: Design
- Status: Active
- Author: masahiro.kasatani

## Overview

kizami は開発ワークフローの2つのポイントで統合される。ドキュメントなしのコミットを警告するローカルの pre-commit フックと、プルリクエストのドキュメントカバレッジをチェックする GitHub Actions ワークフロー（`adr-check.yml`）。どちらもオプトインで `kizami init` によりインストールされる。

## Background

あらゆるドキュメントツールにとって最大のリスクは、開発者がコード変更と同時にドキュメントを書くことを忘れることだ。コミットおよび PR 時の自動リマインダーは、手動プロセスやレビューチェックリストを必要とせずにこのギャップを埋める。どちらの統合ポイントも「ソフトゲート」として設計されており、ブロックではなく警告のみを行う。これにより、もっとも関連性が高いタイミングでリマインダーを表示しながら、摩擦を最小限に抑える。

## Goals / Non-Goals

**Goals:**
- コミット時にドキュメント作成を促す（pre-commit フック）
- PR 時にドキュメントが含まれているか参照されているかをチェックする（CI ワークフロー）
- ドキュメントが本当に不要な PR 向けにエスケープハッチ（PR タイトルの `[skip-doc]`）を提供する
- オプトイン：`kizami init` でユーザーが確認した場合のみインストール
- 既存フックを適切に処理する（上書きせずに内容を出力して手動追記を促す）

**Non-Goals:**
- コミットや PR のハードブロック — どちらの統合も警告のみ
- ADR と PR を構造的にリンクして追跡すること
- PR 説明文への特定の ADR フォーマットの要求

## Design

### Pre-Commit フック

フックスクリプトは POSIX 互換のシェルスクリプト（`templates/pre-commit`）で、`//go:embed` によりバイナリに埋め込まれ、`internal/initializer/hook.go` の `InstallHook` によって `.git/hooks/pre-commit` に書き込まれる。

#### スキップ条件

以下のいずれかが真の場合、フックは 0 で終了する（警告なし）：

1. **ドキュメントファイルがステージ済み**: `git diff --cached --name-only` に `docs/decisions/` 配下のパスが含まれる
2. **ドキュメントのみのコミット**: ステージ済みの全ファイルが `.md` 拡張子を持つ — ドキュメント更新であり新規作成は不要

どちらの条件も満たされず、非 `.md` ファイルがステージされている場合は警告を表示する：

```
⚠️  No ADR found in this commit.
    If this change involves a significant design decision,
    consider running: kizami adr "<title>"
```

フックは**失敗しない**（exit 1 しない）— 常に 0 で終了する。これは意図的な設計だ：ハードブロックは、ドキュメントを必要としない小さなバグ修正やタイポ修正に対して摩擦を生む。

#### 既存フックの処理

`kizami init` 実行時に `.git/hooks/pre-commit` が既に存在する場合、`InstallHook` は上書きしない。代わりにフックスクリプトの内容を stdout に出力し、手動追記を促すメッセージを表示する。既存フックを無言で上書きすると、他のツール（リンター、フォーマッターなど）のフックが壊れてしまうため。

### CI ワークフロー（adr-check.yml）

GitHub Actions ワークフローはすべてのプルリクエストで実行される（`opened`、`edited`、`synchronize` イベント）。

#### チェックロジック

以下のいずれかが真の場合、ワークフローは成功する（exit 0）：

1. **PR タイトルに `[skip-doc]`**: 著者がドキュメント不要を明示している
2. **PR 本文にドキュメントパス**: PR の説明文が `docs/decisions/` または `docs/design/` を参照している — 既存ドキュメントとの関連を示す
3. **変更ファイルにドキュメントファイル**: `git diff --name-only BASE..HEAD` に `^docs/(decisions|design)/.*\.md$` にマッチするパスが含まれる

どれも真でない場合、GitHub Actions の警告アノテーションを出力する：

```
::warning::No document found. Consider adding a decision record to docs/decisions/ or a design doc to docs/design/, or include [skip-doc] in the PR title.
```

警告アノテーションは PR のチェック UI に表示されるが、チェックを「失敗」にはしない。これによりノンブロッキングを維持しながら可視性を確保する。

#### 警告にとどめる理由

ドキュメントカバレッジをハード要件にすると：
- ホットフィックスや緊急 PR がブロックされる
- 設計上の意思決定を含まないリファクタリングや依存関係の更新でノイズが増える
- kizami を段階的に導入しているチームに摩擦が生まれる

警告は重要な変更に対して行動を促すのに十分な可視性を持ちながら、日常的な作業に障壁を作らない。

### フックと CI の関係

2つの統合は補完的かつ独立している：

| | Pre-commit フック | CI ワークフロー |
|---|---|---|
| トリガー | `git commit` | プルリクエスト |
| スコープ | ステージ済みファイル | PR 内の全コミット |
| チェック対象 | `docs/decisions/` のステージ | diff または PR 本文の `docs/(decisions\|design)/` |
| エスケープハッチ | なし（条件未達時は常に警告） | PR タイトルの `[skip-doc]` |
| ブロック | なし | なし |

フックはできるだけ早い段階（コミット時）に問題を捕捉し、CI ワークフローは統合ポイント（PR オープン・更新時）で捕捉する。

## Open Questions

- **docs パスの設定可能化**: フックとワークフローは `docs/decisions/` と `docs/design/` をハードコードしている。異なるパスを使うチームは生成ファイルを手動編集する必要がある。将来的には `kizami init` 時に `kizami.toml` の設定値からパスをテンプレート化できる。
- **ハードブロックモード**: 一部のチームはドキュメントなしの PR を厳格に拒否したい場合がある。ワークフロー内のオプトイン `strict: true` フラグで警告をエラーに変更できる仕組みが考えられる。
- **pre-commit フックの `docs/design/` 対応**: pre-commit フックは `docs/decisions/` のみをチェックし、`docs/design/` はチェックしない。CI ワークフローとの既知の不整合。

## Related Files

- `internal/initializer/hook.go`
- `internal/initializer/templates/pre-commit`
- `internal/initializer/templates/adr-check.yml`
- `internal/initializer/init.go`
- `cmd/init.go`
