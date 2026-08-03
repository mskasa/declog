# `kizami hook pre-tool-use`：Claude CodeのPreToolUseフックによる決定論的な文脈注入

- Date: 2026-08-03
- Type: ADR
- Status: Active
- Author: masahiro.kasatani

## Context

設計ドキュメント（[[agent-context-layer-design]]）は、統治する決定を提示する手段としてマニフェスト（Step 2）とMCPサーバー（Step 3）を挙げているが、両方とも依然としてエージェント自身の判断に依存している：自分の文脈ファイルを読むか、ツールを呼ぶと決めるかである。`kizami hook pre-tool-use` は本フェーズで唯一の決定論的な経路である——統治対象ファイルへのEdit/Writeの直前に、エージェントの判断を介さず無条件で発火する必要があり、他の2つの経路が構造的に埋められないギャップを埋める。

これには、「フック」という一般的な概念ではなく、Claude Codeの実際のフックプロトコルとの統合が必要になる。決めつけずに現行のプロトコルを確認したところ：Claude Codeの `PreToolUse` フックは標準入力でJSON（`Edit`／`Write` の場合は `tool_name`、`tool_input.file_path`、`cwd` など）を受け取り——重要なのは——標準出力に `{"hookSpecificOutput": {"hookEventName": "PreToolUse", "additionalContext": "..."}}` を終了コード0で返すことで、ツール呼び出しを*ブロックせずに*Claudeの文脈にテキストを注入できる。これは本ステップが必要とする、まさにそのプリミティブである：`permissionDecision: "deny"`／`"ask"` も存在するが、編集をブロックしたり中断させたりすることはそもそもここでの目標ではない。設計ドキュメントは、本フェーズが決定の提示であり、編集のゲーティングではないと明示している。

## Decision

`kizami hook pre-tool-use` は標準入力から `PreToolUse` のJSONを読み、`tool_input.file_path` を抽出してリポジトリルート相対パスに解決し、本フェーズの他のすべてのツールが既に使っている同じ `internal/context.Resolve` を呼ぶ。何もマッチしなければ、出力なしで終了コード0——通常時であり、静かで、コストも低い。何かマッチした場合は `additionalContext` を出力して終了コード0とする。deny も ask も一切行わない。本ツールの仕事はただ1つ：ブロックすることが本フェーズの本来の目標に逆行するであろう、まさにその瞬間に、Step 1の文脈を見落とすことが不可能な形にすることである。

注入するテキストは、エージェント向けマニフェスト（Step 2）やMCPツールのレスポンス（Step 3）より意図的に簡潔にする：マッチした決定1件につき1行——slug・1行タイトル・ドリフトフラグ（あれば）——のみとし、本文の代わりに `kizami show <slug>` へのポインタを示す。決定本文はインライン化しない。根拠：マニフェストはセッションごとに1回ロードされ、MCPツールはオンデマンドで呼ばれるが、本フックは統治対象ファイルへの*あらゆる*Edit/Writeで発火し、1セッションで数十回に及ぶこともある。マニフェストの既に簡潔な160文字への切り詰め（[[agent-manifest-sync-format]] 参照）でさえ、この呼び出し頻度では悪化して積み重なる。この頻度で安価であり続けられる形は、ポインタしかない。

失敗は静かに扱う。`kizami hook pre-commit` の既存の規約（gitルートや設定読み込みのエラーで `return nil`）をそのまま踏襲する——本フックは、「今たまたきgitリポジトリの中にいない」といった些細な理由で編集がブロックされたりセッションがエラーになったりする原因になってはならない。

配線はコードではなくドキュメントで行う：本ステップでは `kizami init` は `.claude/settings.json` を管理しない。これを使いたいユーザーは自分で追加する。

```json
{
  "hooks": {
    "PreToolUse": [
      { "matcher": "Edit|Write", "hooks": [{ "type": "command", "command": "kizami hook pre-tool-use" }] }
    ]
  }
}
```

## Consequences

- これは、Agent Context Layerの全フェーズの中で唯一、エージェントの協力なしに決定を提示する経路である——Claude Code自身のツール実行ループから、編集が起きる前に発火する。
- Claude Code固有である（`PreToolUse` のJSON形状、`additionalContext`、`.claude/settings.json` は、MCPのようなエージェント横断の標準ではない）ため、この経路はClaude Codeでのみ役立つ。これに対し、マニフェストとMCPサーバーは、MCP対応の、あるいはCLAUDE.md/AGENTS.mdを読む任意のエージェントで機能する。この非対称性は見落としではなく、受け入れた上でのものである：今日実際に決定論的に機能する唯一の経路の代償である。
- 本フックは、統治対象ファイルへのEdit/Writeごとに1プロセスの起動を追加する。`internal/context.Resolve` のコストはドキュメント数（`internal/config` 自身が想定する数十〜数百）に有界であるため許容範囲だが、マニフェストやMCP経路にはない、編集ごとの実コストである。
- `.claude/settings.json` を自動で配線しないため、導入には手動のステップが必要になる。これは本ステップで解決するのではなく、先送りする（Alternatives参照）。

## Alternatives Considered

**`kizami init` が `.claude/settings.json` のフック項目を自動で書き込む**
手動ステップを排除できるが、`kizami init` は現在 `.claude/` 配下のものに一切触れておらず、ユーザーが既にカスタマイズしている可能性のあるJSONファイル（既存のフック、他の設定）へのマージは、`kizami init` が今日書き込んでいるMarkdown/YAMLテンプレートファイルとは、性質も リスクも明確に異なる操作である。ドキュメント化された手動ステップとして残す。自動配線は将来の候補であり、ここで先回りして解決しない。

**`kizami context` の既定に合わせて、`additionalContext` に決定本文全体を含める**
Step 1のレスポンス形状とは一致するが、本フックの呼び出し頻度（統治対象ファイルの編集ごと。問い合わせごとに1回ではない）では、要約のみという既定でさえここでは高価すぎる——Decision参照。slugのみのポインタは、本フェーズの他の部分の「既定は要約」というルールに対する、頻度に基づいた意図的な例外であり、矛盾ではない。

**決定がファイルを統治している場合、`permissionDecision: "ask"` を使ってユーザーに続行前の確認を求める**
文脈が確実に見られることを保証できるが、情報提示の仕組みを、統治対象ファイルの編集ごとのゲートに変えてしまう。これは設計ドキュメントが本ステップを「決定の提示であり、編集のゲーティングではない」と位置づけていることに直接反し、統治対象ファイルを触ることを測定可能な形で煩わしくしてしまい、低摩擦な導入という目標を損なう。

## Related Files

- `cmd/hook.go`
- `internal/decision/hook.go`
