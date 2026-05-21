---
layout: default
title: Configuration
nav_order: 6
---

# Configuration

kizami can be configured via a `kizami.toml` file in the repository root, or a global config at `~/.config/kizami/config.toml`.

Project-level config (`kizami.toml`) takes precedence over the global config.

[← Back to Documentation](.)

---

## Creating a config file

Run `kizami init` to generate a `kizami.toml` with default values, or create one manually.

---

## All options

### `[documents]`

```toml
[documents]
dirs = ["docs/decisions", "docs/design"]
```

Directories that `kizami list`, `kizami search`, `kizami show`, `kizami blame`, `kizami review`, `kizami status`, and `kizami supersede` scan for documents. Defaults to `[decisions] dir` if not set.

### `[decisions]`

```toml
[decisions]
dir = "docs/decisions"
```

Directory where `kizami adr` saves new ADRs.

### `[design]`

```toml
[design]
dir = "docs/design"
```

Directory where `kizami design` saves new design documents.

### `[audit]`

```toml
[audit]
dirs = ["docs/decisions", "docs/design"]
```

Directories that `kizami audit` checks. Defaults to `[documents] dirs` if not set.

### `[review]`

```toml
[review]
months_threshold = 6
```

Number of months without update before `kizami review` considers a document stale. Defaults to 6. Can also be overridden per run with `kizami review --months N`.

### `[editor]`

```toml
[editor]
command = "code --wait"
```

Editor command used to open documents after creation. Overridden by the `$EDITOR` and `$VISUAL` environment variables.

### `[ai]`

```toml
[ai]
model = "claude-sonnet-4-20250514"
```

Model used by `kizami adr --ai` and `kizami design --ai`. Can be overridden per run with `--model`.

**Anthropic API (default):** set a standard model ID such as `claude-sonnet-4-20250514`. Requires `ANTHROPIC_API_KEY`.

**AWS Bedrock:** set a Bedrock model ID and kizami automatically routes to Bedrock — no extra configuration needed. Supported patterns:

| Pattern | Example |
|---|---|
| Application / Provisioned Inference Profile ARN | `arn:aws:bedrock:ap-northeast-1:123456789012:application-inference-profile/abc123` |
| Cross-region inference profile | `us.anthropic.claude-3-5-sonnet-20241022-v2:0` |
| Provider-prefixed model ID | `anthropic.claude-3-5-sonnet-20241022-v2:0` |

AWS credentials are resolved from the standard environment (`AWS_PROFILE`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, etc.).

You can also force Bedrock explicitly via `[ai] backend = "bedrock"` or the `CLAUDE_CODE_USE_BEDROCK=1` environment variable.

---

## Example `kizami.toml`

```toml
[documents]
dirs = ["docs/decisions", "docs/design"]

[decisions]
dir = "docs/decisions"

[design]
dir = "docs/design"

[audit]
dirs = ["docs/decisions", "docs/design"]

[review]
months_threshold = 6

[editor]
command = "code --wait"

[ai]
model = "claude-sonnet-4-20250514"
```
