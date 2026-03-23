# スケジュール CI ワークフロー

- Date: 2026-03-23
- Type: Design
- Status: Active
- Author: masahiro.kasatani

## Overview

kizami は2つのスケジュール GitHub Actions ワークフローを生成する。`adr-audit.yml` は週次で `kizami audit` を実行し、陳腐化した参照が見つかった場合に GitHub Issue を作成する。`kizami-promote.yml` は `main` へのプッシュごとにドキュメントを `Status: Draft` から `Status: Active` へ自動昇格させる。

## Background

リビングドキュメントには2つの継続的なメンテナンスシグナルが必要だ。ドキュメントが記述するコードから乖離していることを検出すること、そして開発中に書かれたドラフトドキュメントを作業マージ後に正式にアクティブ化することだ。

どちらも個々の開発者が覚えておくべき定期タスクではない。スケジュールまたはイベント駆動の CI ワークフローとしてエンコードすることで、チームの習慣に関係なく一貫して実行されることを保証する。

## Goals / Non-Goals

**Goals:**
- `adr-audit.yml`: 週次スケジュールと手動トリガーで `kizami audit` を実行し、陳腐化した参照が見つかった場合に GitHub Issue を作成し、スパム防止のために Issue を重複排除する
- `kizami-promote.yml`: `main` へのプッシュ時にすべての `Status: Draft` ドキュメントを `Status: Active` に昇格させ、`[skip ci]` 付きボットコミットとして変更をコミットして CI ループを防ぐ
- どちらのワークフローも `kizami init` で生成（オプトイン）
- どちらも自己完結型 — 公開リリースから kizami をインストールし、他のツールを必要としない

**Non-Goals:**
- 陳腐化した参照の自動修正（audit はレポートのみ；修正は人間が判断）
- `main` 以外のブランチでのドキュメント昇格
- GitHub Actions 以外の CI システムへの対応

## Design

### Audit ワークフロー（`adr-audit.yml`）

#### トリガー

- `schedule: cron: '0 0 * * 1'` — 毎週月曜 00:00 UTC に実行
- `workflow_dispatch` — GitHub Actions UI から手動実行可能

#### ステップ

1. **Checkout**（`fetch-depth: 0`）— `kizami audit` がドキュメントセット全体を処理するためにフル履歴が必要
2. **Go のセットアップ**（`actions/setup-go`）
3. **kizami のインストール**（`go install github.com/mskasa/kizami@latest`）— 常に公開リリースを使用し、ローカルの Go ツールチェーンに依存しない
4. **`kizami audit` の実行**（ステップ id: `audit`）— 陳腐化した参照が検出された場合にアウトプット `stale_found=true` を設定する
5. **GitHub Issue の作成**（`steps.audit.outputs.stale_found == 'true'` の場合のみ）`actions/github-script` を使用

#### Issue の重複排除

Issue 作成前に、スクリプトはすべてのオープン Issue を一覧取得し、タイトルに `[kizami audit]` が含まれるものがあるか確認する。重複が存在する場合は新しい Issue を作成しない。これにより、陳腐化した参照が速やかに修正されない場合に、同一の Issue が週次で積み重なるのを防ぐ。

Issue タイトルには現在の日付が含まれる：
```
[kizami audit] Stale file references detected (2026-03-23)
```

Issue 本文には `AUDIT_RESULT` 環境変数経由で `kizami audit` の全出力が含まれる。

#### 必要なパーミッション

ジョブは GitHub Issue を作成するために `issues: write` が必要。ジョブの `permissions` ブロックに明示的に宣言される。

### プロモートワークフロー（`kizami-promote.yml`）

#### トリガー

- `main` への `push` — デフォルトブランチへの全マージで実行

#### マージ時に Draft → Active に昇格する理由

ドキュメントは開発中（意思決定がまだ進行中またはレビュー中）に `Status: Draft` で作成される。フィーチャーブランチが `main` にマージされた時点で、その意思決定は確定したとみなされる。マージ時に自動昇格することで、手動のステータス更新を必要とせずに、ドキュメントのステータスをコードのライフサイクルに合わせられる。

#### ステップ

1. **Checkout** — デフォルトのシャロークローンで十分（git 履歴は不要）
2. **Draft ドキュメントの昇格** — シェルスクリプトが `docs/decisions` と `docs/design` の `.md` ファイルをスキャンし、`- Status: Draft` を含むファイルを `sed` で `- Status: Active` に置換し、ファイルが変更された場合は `changed=1` を設定する
3. **コミット**（`changed == 1` の場合のみ）— `github-actions[bot]` として `docs: auto-promote Draft documents to Active [skip ci]` というメッセージでコミットする。`[skip ci]` によりこのプッシュが再度ワークフローをトリガーするのを防ぐ

#### カスタマイズポイント

ワークフローには2つのカスタマイズポイントを指すインラインコメントが意図的に記載されている：
- `main` ブランチ名（`master` や別のデフォルトブランチを使うリポジトリ向け）
- スキャン対象ディレクトリを列挙する `dirs` 変数

これらはローカルの設定ファイルにワイヤリングせず生成ファイル内の編集可能フィールドとして残す。ワークフローは GitHub Actions 上で実行され、ローカルの config ファイルにアクセスできないため。

#### ボットコミットの識別情報

コミットは以下を使用する：
```
user.name:  github-actions[bot]
user.email: github-actions[bot]@users.noreply.github.com
```

これは標準的な GitHub Actions ボットの識別情報であり、git ログで認識しやすく、個々の開発者への誤帰属を避けられる。

## Open Questions

- **`kizami audit` の終了コード**: 現在 `kizami audit` は陳腐化した参照が見つかっても非ゼロの終了コードを設定しない（ローカルシェルで問題が起きないように）。CI ワークフローは代わりに GitHub Actions のアウトプット変数（`stale_found`）に依存している。将来的な `--strict` フラグでコマンドを CI 用途で非ゼロ終了させることができる。
- **main 以外のブランチでのプロモート**: `release/` ブランチモデルを使うチームもある。プロモートワークフローは `main` のみをターゲットとしており、他のパターンには手動編集が必要。
- **プロモートのスコープ**: ワークフローはスキャン対象の全ディレクトリにわたってすべての Draft ドキュメントを昇格させる。ドキュメント単位のオプトアウトはない。`promote: never` メタデータフィールドで特定のドキュメントを自動昇格から除外できる仕組みが考えられる。

## Related Files

- `internal/initializer/templates/adr-audit.yml`
- `internal/initializer/templates/kizami-promote.yml`
- `internal/initializer/init.go`
- `internal/decision/audit.go`
- `cmd/audit.go`
