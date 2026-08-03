# エージェント文脈層（Agent Context Layer）

- Date: 2026-08-03
- Type: Design
- Status: Active
- Author: masahiro.kasatani

## Overview

kizami は現状、人間が能動的にCLI（`show`・`blame`・`search`）を打った時にのみ価値を発揮する。本設計は重心を「必要な瞬間に決定が自動で現れる」へ移す——特にAIコーディングエージェント向けに。kizami の `## Related Files` によるリンク付けとドリフト検知が最も効くのはそこだからである。本設計では単一のリゾルバパッケージを導入し、その上に4つの利用者（CLIコマンド、エージェントが読むファイルへ同期されるマニフェスト、MCPサーバー、そして最終ステップとしてエージェントによる決定の書き込み経路）を構築する。

## Background

到達性は課題ではない。リポジトリを読むエージェントは既に `docs/decisions/*.md` を直接開ける。課題は別の2箇所にある。

1. **決定が下された瞬間に確実に書かれない。** 現状ADRを書くには、実装作業の最後——最も書く意欲が低いタイミング——で人間が文脈スイッチする必要がある。導入したのに期待した効果が出ていない主因はここにある：記録の蓄積が少なすぎる。
2. **必要な瞬間に確実に提示されない。** ドキュメントが存在していても、対象ファイルを編集する前にエージェントがそれを読むよう促すものが何もない。`kizami show`/`blame` は、人間（あるいはエージェント）がそれを実行しようと考えることに依存している。

より狭い第3の課題として、`kizami audit` と `kizami blame` は現在「関連性」を2通りの異なる方法で計算している——
`search.Blame`（`internal/search/blame.go`）はファイルパスをドキュメント中に全文検索するのに対し、
`decision.CheckHook`（`internal/decision/hook.go`）は `## Related Files` リストを構造的にパースする。エージェント向けAPIは「このファイルを縛る決定は何か」に対して2つの異なる「縛る」の定義で答えることはできない。これは外部に公開する前に統一しておく必要があり、そのため本設計のStep 1として最初に扱う（[[related-files-single-definition]] 参照）。

## Goals / Non-Goals

**Goals:**
- 「あるファイルをどの決定が縛るか」の権威ある定義を1つに統一し、以下の全利用者（CLI・ドキュメント・フック・MCP）で再利用する。
- エージェントが既に答えを知っているケース（まさに今その判断を下した場合）で、決定を書くコストをほぼゼロにする。
- 決定の*存在*を、エージェントが見に行こうと思うかどうかだけに依存させず、見落とし不可能にする。
- 各利用者のトークン/出力コストを有界かつ予測可能に保つ——エージェント本来のタスクを圧迫しないことが、実運用で機能するための前提条件である。

**Non-Goals:**
- 意味的ドリフト検知（コードが決定の*内容*とまだ一致しているか。ファイルの存在有無だけではない）。ロードマップ上の別項目として扱う。本フェーズのリゾルバの `drift` フィールドは、既存の `kizami audit` と同じ存在チェックのみに限定する。
- `kizami review` のchurnベースの陳腐化スコアリング。別のロードマップ項目であり、本設計とは独立。
- `kizami mcp` の読み取り経路内でのLLM/ネットワーク呼び出し。リゾルバはローカルのMarkdown/YAMLのみを読む。これにより読み取り経路を無料に保ち（前回の料金に関する議論を参照）、依存も増やさない。

## Design

```
                    ┌──────────────────────────────────┐
   kizami context ──┤                                  │
   (CLI / CI)       │   internal/context (リゾルバ)      │
                     │                                  │
   agents sync   ────┤   files[] -> 縛る決定群            │
   (Step 2)          │   + ドリフト状態（存在チェックのみ） │
                      │   単一の「関連」定義               │
   kizami mcp     ───┤                                  │
   (Step 3/4)         └──────────────────────────────────┘
```

`internal/context` は本フェーズの土台となる唯一の新規パッケージである。`search.Blame` と `decision.CheckHook` に
重複していた「関連ファイル」ロジックを置き換え（[[related-files-single-definition]] により統合）、以下を追加する。

- マッチングルール：完全一致・ディレクトリ前置（既存の `trailing-slash/` 規約）・glob（新規）。
- JSONシリアライズ可能な結果型。マッチした決定ごとに slug・title・status・`## Decision` セクションのみ（本文全体ではない。トークン予算の根拠は後述）・マッチしたルール・存在ベースのドリフト状態を保持する。
- Superseded な決定を破棄しない。`Status: Superseded by <slug>` の決定がマッチした場合も `supersededBy` フィールド付きで返す。「以前はこうだったが、今はXを見よ」は、そうしなければ撤回済みの決定をエージェントが繰り返しかねないため、価値の高い文脈である。
- `unmatched`（どの決定にもカバーされていない変更ファイル）もマッチ結果と併せて返す。Step 1 では受動的だが、後段の「ここに決定を書くべきでは」というシグナルの種になる。

上に載る4つの利用者は、それぞれBackgroundで述べた2つの課題の異なる半分を解く。

| 利用者 | コマンド | 解く課題 | 信頼性 |
|---|---|---|---|
| CLI / CI | `kizami context <files...> [--json] [--full]` | どちらでもない。共有プリミティブ（PRコメント自動化の基盤にもなる） | N/A |
| エージェント向けマニフェスト | `kizami agents sync` | 提示——CLAUDE.md/AGENTS.md に常時ロードされる安価なポインタ表 | エージェントが自身の文脈ファイルを読むかどうかに依存 |
| MCPサーバー | `kizami mcp` | 提示——オンデマンドかつ精密 | エージェントがツールを呼ぶ判断をするかどうかに依存 |
| ツールフック | `kizami hook pre-tool-use` | 提示——決定論的 | Edit/Write の直前に無条件で発火し、エージェントの判断を介さない |
| 書き込み経路 | `kizami_record_decision`（MCP） | 記述——エージェントが文脈を最も持っている瞬間にコストをほぼゼロに圧縮する | エージェントが呼ぶ判断をするかどうかに依存（フック/マニフェストによるリマインドで緩和） |

トークンコストは後付けではなく設計の第一級の制約である。`kizami context` とMCPの読み取りツールは、既定で
`## Decision` セクションのみを返す（`full` フラグで全文にエスカレーション可能）。マニフェストは決定本文ではなく
1決定1行（パス→決定へのポインタ）のみを列挙する。この制約がなければ、大きな差分で12件の決定がマッチした場合に
エージェント本来のタスク文脈を圧迫してしまう——これは「誰も呼ばないMCPサーバーは、MCPサーバーがないより悪い」
という失敗モードと同根である。

## Implementation Plan

1. **コンテキストリゾルバ**（本ブランチ）— `internal/context`。`Blame`/`CheckHook` の統合、glob対応、
   `kizami context <files...> [--json] [--full]`。
2. **エージェント向けマニフェスト同期** — `kizami agents sync`（マーカー区間のポインタ表をCLAUDE.md/AGENTS.mdに書き込む）
   と CI用の `kizami agents sync --check`。
3. **MCPサーバー** — `kizami mcp` で `kizami_decisions_for_files`・`kizami_search_decisions`・`kizami_get_decision` を
   「動詞ではなく問い」の形で公開する（CLI動詞の1:1移植ではない。詳細は当該ステップと併せて作成するADRを参照）。
4. **エージェントによる決定の記録** — `kizami_record_decision`（書き込み系。常に `Status: Draft`、新規ファイル作成のみ、
   `kizami mcp --allow-write` でオプトイン）と `kizami hook pre-tool-use`（Edit/Write直前の決定論的な文脈注入）。

各ステップは `feature/agent-context-layer` へのPRとして個別に着地する。本ドキュメントのRelated Filesは、
ADR Update Policy に従い、各ステップのファイルが着地するたびに追記する。

## Open Questions

- **globの実装方法。** `path.Match` は `**` に対応していない。Step 1 の結論は [[related-files-single-definition]] に記す。
- **MCPのトランスポート/SDK選定。** 公式の `github.com/modelcontextprotocol/go-sdk` を使うか、小規模な自前のstdio
  JSON-RPC実装にするか。Step 3 のADRに持ち越す——MCPサーバーが存在するまでは不要な判断のため。
- **トークン予算の既定値。** `## Decision` のみへの切り詰めに、さらなる上限（例：1レスポンスあたりの決定件数上限）が
  必要かどうかは、Step 3 リリース後の実利用を見て再検討する。

## Related Files

- `internal/context/`（新規 — Step 1）
- `cmd/context.go`（新規 — Step 1）
- `internal/decision/match.go`（新規 — Step 1）
- `internal/decision/decision.go`（Step 1：`SupersededBy`）
- `internal/decision/audit.go`
- `internal/decision/hook.go`
- `internal/search/blame.go`
