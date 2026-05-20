# Add AWS Bedrock as an AI backend alternative to Anthropic API

- Date: 2026-05-20
- Status: Active
- Author: masahiro.kasatani

## Context

`kizami adr --ai` and `kizami design --ai` previously called the Anthropic API directly, requiring `ANTHROPIC_API_KEY`. Teams using AWS Bedrock as their AI gateway (e.g. via AWS IAM roles, organization-managed inference profiles, or Application Inference Profiles) could not use these commands without a personal Anthropic API key.

## Decision

Add AWS Bedrock as an alternative AI backend. Backend selection follows this priority:

1. `CLAUDE_CODE_USE_BEDROCK=1` environment variable → Bedrock
2. `[ai] backend = "bedrock"` in `kizami.toml` → Bedrock
3. Default → Anthropic API

When Bedrock is selected, AWS credentials are resolved from the standard environment (`AWS_PROFILE`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, etc.) via `aws-sdk-go-v2`. The model ID (from `--model` flag or `[ai] model` in config) is passed directly to Bedrock's `InvokeModel` API, supporting both standard model IDs and Application Inference Profile ARNs.

The request body for Bedrock Claude models uses `anthropic_version: "bedrock-2023-05-31"` instead of the `x-api-key` header required by the direct Anthropic API.

## Consequences

- Teams on AWS Bedrock can use `kizami adr --ai` without an Anthropic API key.
- `aws-sdk-go-v2` is added as a dependency, increasing binary size.
- Error messages on `--ai` flag now mention both backends, guiding users to the appropriate solution.

## Alternatives Considered

**Re-implement AWS Signature V4 without the SDK**: Avoids the dependency but is complex and error-prone. The SDK is the right tool here.

**Separate `--bedrock` flag**: More explicit, but reuses Claude Code's existing `CLAUDE_CODE_USE_BEDROCK` convention so users on teams already using that env var get it for free.

## Related Files

- internal/ai/bedrock.go
- internal/ai/draft.go
- internal/config/config.go
- cmd/log.go
