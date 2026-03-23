# プレーンテキストMarkdownのみを使用する

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

設計上の意思決定はどこかに保存する必要がある。選択肢はデータベース・構造化ファイル・プレーンテキストなど多岐にわたる。
保存フォーマットは、ポータビリティ・Gitのdiff品質・記録を読むために必要なツールに影響する。
declogはGitをすでに使用しているソフトウェアプロジェクト向けに設計されている。

## Decision

すべての意思決定記録を`docs/decisions/`配下のプレーンMarkdownファイルとして保存する。
データベース・バイナリフォーマット・追加のメタデータファイルは使用しない。

## Consequences

- ファイルはツールなしで人間が読める——`cat`・任意のテキストエディタ・GitHubのWeb UIすべてで閲覧可能
- Gitとの親和性が高い：行単位のdiffが意味を持ち、blameも機能し、履歴が明確
- ポータブル：意思決定記録はリポジトリと共に移動し、オフラインでもアクセス可能
- マイグレーション不要：今日書いた記録はどんな将来の環境でも読める
- 検索はファイルシステムとripgrepに委ねる——クエリ言語が不要

## Alternatives Considered

- **SQLiteデータベース：** 構造化クエリが可能だが、ポータビリティを損ないバイナリdiffになる
- **JSONやYAMLファイル：** 機械可読だが、直接書いたり読んだりするには不便
- **専用ADRサービス（外部SaaSなど）：** 一元管理できるが外部依存が生まれ、意思決定記録がコードベースから分離される

## Related Files

- `internal/decision/decision.go`
- `internal/decision/generate.go`
- `internal/template/template.go`
