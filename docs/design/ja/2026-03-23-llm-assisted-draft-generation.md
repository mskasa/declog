# LLMによるドラフト自動生成

- Date: 2026-03-23
- Type: Design
- Status: Active
- Author: masahiro.kasatani

## Overview

`kizami adr --ai` および `kizami design --ai` は、Anthropic API を使って現在の git コンテキスト（変更ファイルとステージ済み diff）からドキュメントのドラフトを生成する。コード変更と同時にリビングドキュメントを書く際の摩擦を減らすことが目的。

## Background

有用な ADR や設計書を書くには、コード変更に暗黙的に含まれているコンテキスト（どのファイルが変更されたか、diff が何を示しているか、解決しようとした問題は何か）を正確に捉える必要がある。しかしこのコンテキストを事後に記憶から再構築するのは困難であり、ドキュメントの質を下げ、記録の習慣を妨げる要因になる。

kizami はドキュメント作成時にステージ済み diff と変更ファイル一覧を LLM に渡し、意味のある初稿を生成する。開発者は空のテンプレートから書き始めるのではなく、生成された初稿を編集して完成させるだけでよい。

## Goals / Non-Goals

**Goals:**
- git コンテキストから本文セクション（Context/Decision/Consequences または Overview/Background/Design 等）を生成する
- メタデータヘッダー（Date, Type, Status, Author, Supersedes）は常に決定論的に生成し、LLM には生成させない
- `--dry-run` フラグでプロンプトを API 送信前に確認できるようにする
- `--model` フラグと `[ai] model` config キーでデフォルトモデルを上書き可能にする
- ADR（`kizami adr --ai`）と設計書（`kizami design --ai`）の両方に対応する

**Non-Goals:**
- ストリーミング出力 — レスポンスを全件受信してからファイルに書き込む
- 人間のレビューなしの自動保存 — 生成後は必ずエディタを開く
- メタデータフィールド（日付、ステータス、著者）の生成 — これらは常にハードコード
- Anthropic 以外の LLM プロバイダーへの対応

## Design

### フロー

```
kizami adr --ai "<title>"
  │
  ├── 1. promptSimilar()         — 非 AI フローと同様；類似ドキュメントを確認
  ├── 2. GatherInput()           — git から変更ファイルとステージ済み diff を収集
  ├── 3. BuildPrompt()           — プロンプト文字列を構築
  ├── 4. [--dry-run] DryRun()    — プロンプトを表示し確認を求める
  ├── 5. GenerateDraft()         — Anthropic Messages API を呼び出す
  ├── 6. RenderHeader()          — 決定論的なメタデータヘッダーを生成
  ├── 7. CreateFromDraft()       — header + LLM 本文を YYYY-MM-DD-slug.md に書き込む
  └── 8. openEditor()            — エディタでファイルを開く
```

### コンテキスト収集（`internal/ai/prompt.go`）

`GatherInput(dir, title string)` が収集するもの：
- **変更ファイル**: `git diff --name-only` と `git diff --staged --name-only` の和集合（重複除去）
- **ステージ済み diff**: `git diff --staged` の出力。トークン予算を考慮して `DiffLimit = 2000` 文字で切り捨て

diff の切り捨ては、実装をシンプルかつ決定論的に保つために行う。2000 文字という上限は、ドラフト品質と API コストのバランスを考慮して設定した。

### プロンプト構築

ドキュメント種別ごとに専用のプロンプトビルダーが存在する：

- `BuildPrompt(input)` — `## Context`、`## Decision`、`## Consequences`、`## Related Files` セクションを Markdown で生成するよう指示する
- `BuildDesignPrompt(input)` — `## Overview`、`## Background`、`## Goals / Non-Goals`、`## Design`、`## Implementation Plan`、`## Open Questions`、`## Related Files` セクションを Markdown で生成するよう指示する

どちらのプロンプトも：
- 「Markdown のみを出力すること。説明や前置きは不要。」と明示することで、余分なテキストを抑制する
- 変更ファイル一覧と切り捨てた diff をコンテキストとして渡す
- メタデータヘッダーの生成は LLM に求めない

### ヘッダーと本文の分離

メタデータヘッダーは `internal/template/template.go` の `RenderHeader` / `RenderDesignHeader` によって常に決定論的に生成される。LLM が生成するのは `## Context` または `## Overview` 以降の本文セクションのみ。

`internal/decision/generate.go` の `CreateFromDraft` / `CreateDesignFromDraft` は以下のように連結する：

```
header  =  RenderHeader(title, author, supersededSlug)
content =  header + "\n" + draft
```

これにより、Date、Status（`Draft`）、Author、Supersedes がハルシネーションされることはない。

### API 呼び出し（`internal/ai/draft.go`）

`GenerateDraft(prompt, model, apiKey string)` は `POST https://api.anthropic.com/v1/messages` を呼び出す：
- `model`: `ResolveModel(flagModel, cfg)` で解決（フラグ > config > `DefaultModel`）
- `max_tokens`: 2048
- プロンプトを含む単一の `user` メッセージ

`ANTHROPIC_API_KEY` 環境変数が必須。未設定の場合はネットワーク呼び出しを行う前にわかりやすいエラーメッセージを返す。

### モデル解決の優先順位

高い順：
1. コマンドラインの `--model` フラグ
2. `kizami.toml` または `~/.config/kizami/config.toml` の `[ai] model`
3. ハードコードされたデフォルト: `claude-sonnet-4-20250514`（`config.DefaultModel`）

### `--dry-run` フラグ

`DryRun(prompt, r, w)` はプロンプト全文を stdout に表示し、API 呼び出し前に `y/n` の確認を求める。プロンプトの品質確認や、API 料金を発生させずにトークン数を見積もる際に有用。

## Open Questions

- **diff の切り捨て**: 現在の 2000 文字ハードカットは diff のハンク途中で切れる可能性があり、大規模な変更セットではドラフト品質が低下することがある。ハンク単位での切り捨てやファイル優先度付きの切り捨てなど、よりスマートな方法を検討できる。
- **未ステージの diff**: 現在はステージ済み diff のみを送信している。未ステージの変更を含めることでコンテキストが改善する可能性があるが、無関係な編集が含まれるリスクもある。
- **トークン使用量のフィードバック**: API レスポンスにはトークン数が含まれているが、kizami は現在それを表示していない。`--dry-run` 時に表示することで、開発者がワークフローを調整しやすくなる。

## Related Files

- `internal/ai/draft.go`
- `internal/ai/prompt.go`
- `internal/decision/generate.go`
- `internal/template/template.go`
- `cmd/log.go`
- `internal/config/config.go`
