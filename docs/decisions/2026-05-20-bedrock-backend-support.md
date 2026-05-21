# Add AWS Bedrock as an AI backend alternative to Anthropic API

- Date: 2026-05-20
- Status: Active
- Author: masahiro.kasatani

## Context

`kizami adr --ai` and `kizami design --ai` previously called the Anthropic API directly, requiring `ANTHROPIC_API_KEY`. Teams using AWS Bedrock as their AI gateway (e.g. via AWS IAM roles, organization-managed inference profiles, or Application Inference Profiles) could not use these commands without a personal Anthropic API key.

## Decision

Add AWS Bedrock as an alternative AI backend. Backend selection follows this priority:

1. `CLAUDE_CODE_USE_BEDROCK=1` environment variable → Bedrock (Claude Code compatibility)
2. `[ai] backend = "bedrock"` in `kizami.toml` → Bedrock (explicit opt-in)
3. Model ID auto-detection → Bedrock if the model matches a known Bedrock pattern
4. Default → Anthropic API

Recognised Bedrock model patterns (`IsBedrockModel` in `internal/config/config.go`):
- ARN: `arn:aws:bedrock:...` (Application / Provisioned Inference Profile)
- Cross-region inference profile prefix: `us.`, `eu.`, `ap.`
- Provider-prefixed standard model ID: `anthropic.`, `amazon.`, `meta.`, `mistral.`, `cohere.`, `ai21.`, `stability.`

This means the most common path for an OSS user is simply:

```toml
# kizami.toml
[ai]
model = "anthropic.claude-3-5-sonnet-20241022-v2:0"
```

No extra env var or `backend` key is needed; kizami infers Bedrock from the model ID.

When Bedrock is selected, AWS credentials are resolved from the standard environment (`AWS_PROFILE`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, etc.) via `aws-sdk-go-v2`. The model ID (from `--model` flag or `[ai] model` in config) is passed directly to Bedrock's `InvokeModel` API, supporting both standard model IDs and Application Inference Profile ARNs.

The request body for Bedrock Claude models uses `anthropic_version: "bedrock-2023-05-31"` instead of the `x-api-key` header required by the direct Anthropic API.

## Consequences

- Teams on AWS Bedrock can use `kizami adr --ai` by setting a Bedrock model ID in `kizami.toml`, without any additional configuration.
- `aws-sdk-go-v2` is added as a dependency, increasing binary size.
- Error messages on `--ai` flag guide users to both backends.

## Alternatives Considered

**Re-implement AWS Signature V4 without the SDK**: Avoids the dependency but is complex and error-prone. The SDK is the right tool here.

**Separate `--bedrock` flag**: More explicit, but adding auto-detection from model ID is lower friction — users only need to set the model they already know, not learn a kizami-specific flag.

**Require explicit `[ai] backend = "bedrock"`**: Works, but forces users to set two config keys (`model` and `backend`) when the model ID already unambiguously identifies the backend.

## Related Files

- internal/ai/bedrock.go
- internal/ai/draft.go
- internal/config/config.go
- cmd/log.go
