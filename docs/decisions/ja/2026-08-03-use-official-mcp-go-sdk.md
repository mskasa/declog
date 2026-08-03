# `kizami mcp` に公式MCP Go SDKを採用する

- Date: 2026-08-03
- Type: ADR
- Status: Active
- Author: masahiro.kasatani

## Context

Agent Context LayerのStep 3（[[agent-context-layer-design]]）では、`internal/context` のリゾルバをstdio経由で公開するMCPサーバー `kizami mcp` が必要になる。設計ドキュメントではトランスポート／SDKの選定を未決の論点として残していた：公式の `github.com/modelcontextprotocol/go-sdk` を使うか、`use-go-over-shell-script` の一般方針（標準ライブラリに近い労力で済む場合は依存を避ける）に沿って、小さな自前のstdio JSON-RPC実装にするか、である。

決めつけずに実際のSDKの状態を確認したところ：`github.com/modelcontextprotocol/go-sdk` は明示的なAPI互換性保証付きでv1.0.0に達しており、Googleとの協業で保守されていて、本稿執筆時点でv1.7.0（2024-11-05から2026-07-28まで複数のMCPプロトコル仕様バージョンに対応）である。`mcp.NewServer`、Goの構造体タグからJSON Schemaを導出する汎用の `mcp.AddTool[In, Out]`、既製の `mcp.StdioTransport` を提供する。

導入は無料ではない：`go get` はSDK本体に加えて `github.com/google/jsonschema-go` と複数の推移的依存（`segmentio/encoding`、`golang.org/x/oauth2`、`golang.org/x/sync`、`golang.org/x/time`、`yosida95/uritemplate`）を引き込み、SDKはGo 1.25以上を要求するため、本モジュールの `go` ディレクティブが1.24から上がる。依存を最小限にするという技術スタック方針を掲げるプロジェクトにとって、これは実質的なコストであり、既定選択ではなく意図的な決定であるべきだ。

## Decision

公式SDKを使う。決め手はコード量（自前のstdio JSON-RPC実装自体は実際数百行で済む）ではなく、プロトコル面である：MCPのハンドシェイク、ケーパビリティネゴシエーション、JSON-RPCフレーミングは仕様バージョンをまたいで動き続ける対象であり、SDKの役目はまさにその変動を吸収し、`kizami mcp` 自身のコードを本来のドメインロジック（決定の解決）に集中させることにある。これはCLIのパース処理を自前で書く代わりにcobraを選んだ判断（`use-cobra-for-cli-framework`）と同じ理由付けである：プロトコル自身のエコシステムが保守する事実上の標準は、代替が「自分たちで定義したわけではない仕様を再実装し、再保守し続けること」であるなら、依存として受け入れる価値がある。

その結果としてGoバージョンの下限は1.25に上がるが、これは自前実装に留まることで回避するのではなく、承知の上で受け入れる。`.mise.toml`、`CONTRIBUTING.md`、本リポジトリのCLAUDE.mdの技術スタック行を合わせて更新する。CIとリリースワークフローは既に `go.mod` からGoバージョンを読む（`go-version-file: go.mod`）ため、自動的に追従する。

## Consequences

- `kizami mcp` は、kizami自身がプロトコルを追跡する必要なく、upstreamの修正や将来のMCP仕様対応の恩恵を受けられる。
- 依存ツリーが7モジュール増える（SDK＋`jsonschema-go`＋さらに5つの推移的依存）。SDKと（スキーマタグに直接使う場合の）`jsonschema-go` 以外はすべて `// indirect`。
- コントリビューターと `go install` 利用者は今後Go 1.25以上が必要になる（多くの環境では `go.mod` の `go` ディレクティブ経由のGoツールチェイン自動ダウンロードで対応可能）。
- ツールの入出力スキーマは、手書きのJSON Schemaドキュメントではなく、`json`／`jsonschema` タグ付きのGo構造体として宣言する（[[mcp-tools-as-questions-not-verbs]] 参照）。

## Alternatives Considered

**自前のstdio JSON-RPC 2.0実装**
新規依存もGoバージョンの引き上げも不要。しかしMCPは単なるJSON-RPCフレーミングではなく、初期化ハンドシェイク、ケーパビリティネゴシエーション、ツール定義のスキーマ形式を含み、これらすべてを自前で実装し、将来の仕様改訂に追従させ続ける必要がある。SDKが公式であり、安定版（v1.0以上の互換性保証）で、活発に保守されていることを踏まえると、これを自前で再実装することは、kizami本来の価値（決定データと解決ロジックであり、トランスポートではない）に見合わないプロトコル保守コストを無期限に負うことになる。

**依存がより軽い、あるいはpre-1.0の代替を待つ**
実質的に軽量な公式の代替は存在しない。非公式のコミュニティSDKは、同等かそれ以上の依存コストに対して、より大きなリスク（レビューの薄さ、互換性保証の欠如）を伴う。

## Related Files

- `go.mod`
- `internal/mcp/`
- `cmd/mcp.go`
- `.mise.toml`
- `CONTRIBUTING.md`
