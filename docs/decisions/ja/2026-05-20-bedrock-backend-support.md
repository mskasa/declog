# AIバックエンドとしてAWS BedrockをAnthropicの代替に追加する

- Date: 2026-05-20
- Status: Active
- Author: masahiro.kasatani

## Context

`kizami adr --ai` と `kizami design --ai` はこれまでAnthropicのAPIを直接呼び出しており、`ANTHROPIC_API_KEY` が必須だった。AWS BedrockをAIゲートウェイとして利用しているチーム（IAMロール、組織管理のApplication Inference Profileなど）は、個人のAnthropicのAPIキーなしにはこれらのコマンドを使用できなかった。

## Decision

AWS Bedrockを代替AIバックエンドとして追加する。バックエンドの選択は以下の優先順位に従う：

1. 環境変数 `CLAUDE_CODE_USE_BEDROCK=1` → Bedrock（Claude Code互換）
2. `kizami.toml` の `[ai] backend = "bedrock"` → Bedrock（明示的な指定）
3. モデルIDの自動検出 → Bedrockパターンに一致する場合はBedrock
4. デフォルト → Anthropic API

Bedrockモデルとして認識されるパターン（`internal/config/config.go` の `IsBedrockModel`）：
- ARN: `arn:aws:bedrock:...`（Application / Provisioned Inference Profile）
- クロスリージョン推論プロファイルのプレフィックス: `us.`、`eu.`、`ap.`
- プロバイダープレフィックス付き標準モデルID: `anthropic.`、`amazon.`、`meta.`、`mistral.`、`cohere.`、`ai21.`、`stability.`

これにより、OSSユーザーの最も一般的な設定方法は次のとおり：

```toml
# kizami.toml
[ai]
model = "anthropic.claude-3-5-sonnet-20241022-v2:0"
```

追加の環境変数や `backend` キーは不要。kizamiがモデルIDからBedrockを自動判別する。

Bedrockが選択された場合、AWSクレデンシャルは `aws-sdk-go-v2` を通じて標準の環境変数（`AWS_PROFILE`、`AWS_REGION`、`AWS_ACCESS_KEY_ID` など）から解決される。モデルID（`--model` フラグまたはconfigの `[ai] model`）はBedrockの `InvokeModel` APIにそのまま渡され、標準のモデルIDとApplication Inference Profile ARNの両方に対応する。

BedrockのClaudeモデル向けリクエストボディでは、Anthropic API直接呼び出し時の `x-api-key` ヘッダーの代わりに `anthropic_version: "bedrock-2023-05-31"` を使用する。

## Consequences

- AWSを利用するチームがAnthropicのAPIキーなしで `kizami.toml` にBedrockのモデルIDを設定するだけで `kizami adr --ai` を使用できる。
- `aws-sdk-go-v2` が依存として追加されバイナリサイズが増加する。
- `--ai` フラグ利用時のエラーメッセージが両バックエンドへの対応方法を案内する。

## Alternatives Considered

**SDKを使わずにAWS Signature V4を実装する**: 依存を回避できるが複雑でエラーが起きやすい。SDKが適切なツール。

**専用の `--bedrock` フラグ**: より明示的だが、モデルIDからの自動検出のほうが設定の手間が少ない。ユーザーは既に知っているモデルIDを設定するだけでよく、kizami固有のフラグを覚える必要がない。

**`[ai] backend = "bedrock"` の明示的な指定を必須にする**: 動作するが、モデルIDが既にバックエンドを一意に特定できるにもかかわらず、ユーザーに `model` と `backend` の2つのキーを設定させることになる。

## Related Files

- internal/ai/bedrock.go
- internal/ai/draft.go
- internal/config/config.go
- cmd/log.go
