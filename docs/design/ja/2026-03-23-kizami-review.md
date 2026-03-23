# kizami review

- Date: 2026-03-23
- Type: Design
- Status: Active
- Author: masahiro.kasatani

## Overview

`kizami review` は、git のコミット履歴を「最終更新日時」の情報源として、設定可能な期間以上更新されていない Active なドキュメントを一覧表示する。

## Background

リビングドキュメントは正確さを保ち続けることで初めて価値を持つ。2年前の意思決定を記述した ADR や設計書は、現在のシステムを反映していない可能性がある。陳腐化したドキュメントを特定する手段がなければ、チームはドキュメントを一度書いて放置しがちになり、ないよりも悪い誤った参照が残り続ける。

`kizami review` は、git がすでに追跡している情報以外に追加のメタデータを必要とせず、見直しが必要なドキュメントを定期的に浮かび上がらせる低コストな手段を提供する。

## Goals / Non-Goals

**Goals:**
- 最終 git コミットが設定可能な閾値（デフォルト: 6ヶ月）より古い Active ドキュメントをレポートする
- 閾値は `kizami.toml` の `[review] months_threshold` から読み込み、`--months` フラグで実行単位に上書き可能にする
- `documents.dirs` に設定された全ディレクトリをスキャンする
- 陳腐化したドキュメントごとに slug/タイトル、最終更新日、経過月数を表示する
- 「更新、Inactive 化、または Supersede を検討してください」という具体的な提案を出力する

**Non-Goals:**
- 陳腐化したドキュメントの自動更新やクローズ
- 通知や GitHub Issue の作成（それは `kizami audit` の責務）
- ドキュメント内へのレビュー履歴の記録

## Design

### 陳腐化の定義

以下の**すべて**の条件を満たすドキュメントを陳腐化とみなす：
1. `Status` フィールドが `Active`（大文字小文字を区別しない）
2. ファイルに触れた最新の git コミットが `now - months` より古い

Draft、Inactive、Superseded のドキュメントは除外する：作業中または既知の陳腐化ドキュメントであるため。

### git を通じた「最終更新日時」の取得

`LastUpdated(path string) (time.Time, error)` は以下を実行する：

```
git log -1 --format=%ci -- <path>
```

ファイルに触れた最新コミットのコミッター日時を返す。ファイルシステムの mtime ではなく git 履歴を使う理由は、mtime はチェックアウト時に変更されて無意味であり、git コミット日時こそがファイルが実際に変更された時刻を反映しているため。

git 履歴がないファイル（未コミットなど）については `LastUpdated` がエラーを返し、`FindStale` はそのドキュメントを静かにスキップする。

### 閾値の計算

`FindStale` はカットオフを `now.AddDate(0, -months, 0)` として計算する。`LastUpdated` がこのカットオフより前であるドキュメントを陳腐化とみなす。

`MonthsAgo(t, now time.Time) int` は表示用の経過月数を計算する。月内の日付まで考慮してオフバイワンを防ぐ（例：4月15日は5月14日より前であれば経過1ヶ月ではなく0ヶ月）。

### 閾値の解決優先順位

高い順：
1. コマンドラインで明示的に渡された `--months` フラグ
2. `kizami.toml` の `[review] months_threshold`（0より大きい場合）
3. ハードコードされたデフォルト: `6`

`!cmd.Flags().Changed("months")` チェックにより、「フラグが渡されていない」と「フラグがデフォルト値で渡された」を区別し、フラグ省略時に config ファイルの値が有効になるようにする。

### 複数ディレクトリのサポート

`kizami review` は `documents.dirs`（`cmd/root.go` の `documentDirs(root, cfg)` で解決）に列挙された全ディレクトリをイテレートし、各ディレクトリから陳腐化ドキュメントを収集してまとめてレポートする。これにより `docs/decisions/` と `docs/design/` の両方を1回の実行でカバーできる。

### テスタビリティ

`FindStale` は `LastUpdated` を直接呼び出す代わりに、`lastUpdatedFn func(string) (time.Time, error)` パラメータを受け取る。テストでは、制御されたタイムスタンプを返すフェイク関数を注入することで、実際の git リポジトリなしに陳腐化ロジックをテストできる。

## Open Questions

- **ドキュメントメタデータへのレビュータイムスタンプ記録**: 現時点では「このドキュメントを YYYY-MM-DD に確認して正確であることを確認した」という情報を git コミットなしに記録する手段がない。`Reviewed:` メタデータフィールドがあれば、コミット履歴を超えた陳腐化シグナルを補完できる。
- **安定した意思決定に対するノイズ**: 一部の ADR（例：「Go を使う」）は意図的に長命である。6ヶ月ごとに陳腐化フラグが立つのはノイズになる。個別のドキュメントで `review: never` アノテーションを設定することで review レポートから除外できる仕組みが考えられる。

## Related Files

- `internal/decision/review.go`
- `cmd/review.go`
- `internal/config/config.go`
