---
layout: default
title: 設定
parent: 日本語
nav_order: 3
---

# 設定

kizamiはリポジトリルートの `kizami.toml`、またはグローバル設定 `~/.config/kizami/config.toml` で設定できます。

プロジェクトレベルの設定（`kizami.toml`）がグローバル設定より優先されます。

[← ドキュメントトップへ](.)

---

## 設定ファイルの作成

`kizami init` を実行するとデフォルト値入りの `kizami.toml` が生成されます。手動で作成することも可能です。

---

## 全オプション

### `[documents]`

```toml
[documents]
dirs = ["docs/decisions", "docs/design"]
```

`kizami list`・`kizami search`・`kizami show`・`kizami blame`・`kizami review`・`kizami status`・`kizami supersede` がドキュメントを検索するディレクトリです。設定しない場合は `[decisions] dir` が使われます。

### `[decisions]`

```toml
[decisions]
dir = "docs/decisions"
```

`kizami adr` が新しいADRを保存するディレクトリです。

### `[design]`

```toml
[design]
dir = "docs/design"
```

`kizami design` が新しい設計ドキュメントを保存するディレクトリです。

### `[audit]`

```toml
[audit]
dirs = ["docs/decisions", "docs/design"]
```

`kizami audit` がチェックするディレクトリです。設定しない場合は `[documents] dirs` が使われます。

### `[review]`

```toml
[review]
months_threshold = 6
```

`kizami review` が陳腐化とみなすまでの月数です。デフォルトは6ヶ月。`kizami review --months N` でコマンド実行時に上書きできます。

### `[editor]`

```toml
[editor]
command = "code --wait"
```

ドキュメント作成後に開くエディタコマンドです。`$EDITOR`・`$VISUAL` 環境変数が優先されます。

### `[ai]`

```toml
[ai]
model = "claude-sonnet-4-20250514"
```

`kizami adr --ai` / `kizami design --ai` で使用するモデルです。`--model` フラグでコマンド実行時に上書きできます。

**Anthropic API（デフォルト）：** `claude-sonnet-4-20250514` などの標準モデルIDを指定します。`ANTHROPIC_API_KEY` が必要です。

**AWS Bedrock：** BedrockのモデルIDを設定するだけで、kizamiが自動的にBedrockへルーティングします。追加設定は不要です。対応パターン：

| パターン | 例 |
|---|---|
| Application / Provisioned Inference Profile ARN | `arn:aws:bedrock:ap-northeast-1:123456789012:application-inference-profile/abc123` |
| クロスリージョン推論プロファイル | `us.anthropic.claude-3-5-sonnet-20241022-v2:0` |
| プロバイダープレフィックス付きモデルID | `anthropic.claude-3-5-sonnet-20241022-v2:0` |

AWSクレデンシャルは標準の環境変数（`AWS_PROFILE`、`AWS_REGION`、`AWS_ACCESS_KEY_ID` など）から解決されます。

明示的にBedrockを指定したい場合は `[ai] backend = "bedrock"` または環境変数 `CLAUDE_CODE_USE_BEDROCK=1` も使用できます。

---

## `kizami.toml` の例

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
