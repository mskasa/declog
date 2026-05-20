# AIバックエンドとしてAWS BedrockをAnthropicの代替に追加する

- Date: 2026-05-20
- Status: Active
- Author: masahiro.kasatani

## Context

`kizami adr --ai` と `kizami design --ai` はこれまでAnthropicのAPIを直接呼び出しており、`ANTHROPIC_API_KEY` が必須だった。AWS BedrockをAIゲートウェイとして利用しているチーム（IAMロール、組織管理のApplication Inference Profileなど）は、個人のAnthropicのAPIキーなしにはこれらのコマンドを使用できなかった。

## Decision

AWS Bedrockを代替AIバックエンドとして追加する。バックエンドの選択は以下の優先順位に従う：

1. 環境変数 `CLAUDE_CODE_USE_BEDROCK=1` → Bedrock
2. `kizami.toml` の `[ai] backend = "bedrock"` → Bedrock
3. デフォルト → Anthropic API

Bedrockが選択された場合、AWSクレデンシャルは `aws-sdk-go-v2` を通じて標準の環境変数（`AWS_PROFILE`、`AWS_REGION`、`AWS_ACCESS_KEY_ID` など）から解決される。モデルID（`--model` フラグまたはconfigの `[ai] model`）はBedrockの `InvokeModel` APIにそのまま渡され、標準のモデルIDとApplication Inference Profile ARNの両方に対応する。

BedrockのClaudeモデル向けリクエストボディでは、Anthropic API直接呼び出し時の `x-api-key` ヘッダーの代わりに `anthropic_version: "bedrock-2023-05-31"` を使用する。

## Consequences

- AWSを利用するチームがAnthropicのAPIキーなしで `kizami adr --ai` を使用できる。
- `aws-sdk-go-v2` が依存として追加されバイナリサイズが増加する。
- `--ai` フラグ利用時のエラーメッセージが両バックエンドに言及し、ユーザーを適切な解決策に導く。

## Alternatives Considered

**SDKを使わずにAWS Signature V4を実装する**: 依存を回避できるが複雑でエラーが起きやすい。SDKが適切なツール。

**専用の `--bedrock` フラグ**: より明示的だが、Claude Codeが既に使用している `CLAUDE_CODE_USE_BEDROCK` の慣例を再利用することで、そのenv varを設定済みのチームが自動的に恩恵を受けられる。

## Related Files

- internal/ai/bedrock.go
- internal/ai/draft.go
- internal/config/config.go
- cmd/log.go
