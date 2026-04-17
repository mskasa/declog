# Markdown以外のドキュメントへのサイドカーファイルサポート

- Date: 2026-04-17
- Status: Active
- Author: masahiro.kasatani

## Context

これまでkizamiが管理できるのは Markdown ドキュメント（`- Status:` と `## Related Files` を含む `.md` ファイル）のみでした。
CSVテストマトリクス・OpenAPI仕様書・SQLスキーマ・画像といった非Markdownファイルは、別途Markdownドキュメントの `## Related Files` に列挙することで間接的にしか追跡できませんでした。

この間接方式では、橋渡し役のMarkdownドキュメントをわざわざ作る手間が生じます。
目的は、ファイル自体を改変することなく、任意の拡張子のファイルをkizamiのファーストクラスの管理対象にすることです。

## Decision

`.kizami` サイドカーファイルを導入します。
サイドカーは管理対象ファイルと同じ場所に置く小さなYAMLファイルで、`<ファイル名>.kizami` という名前にします。
kizamiはサイドカーをドキュメントとして扱い、元ファイルを追跡対象のアーティファクトとして扱います。

サイドカーフォーマット:

```yaml
title: ユーザーフローのテストマトリクス
date: 2026-04-17
author: masahiro.kasatani
related:
  - tests/user_flow_test.go
```

主な設計上の判断:

- **`status` フィールドなし**: サイドカーはレビュー中の意思決定ではなく、事実としての関係性を表します。
  常にActiveとして扱われ、`kizami audit` の対象に常に含まれます。
- **`date` = 作成日**: Markdown ADRと統一します。更新履歴は git log で追跡します。
- **外部YAMLライブラリなし**: フォーマットが単純なので行ごとにパースでき、依存ライブラリを追加しません。
- **スラグ = 管理対象ファイル名**: `test_matrix.csv.kizami` のスラグは `test_matrix.csv` となり、
  `kizami show test_matrix.csv` で自然に検索できます。
- **ファイルごとにサイドカーを作成（プロジェクト1ファイルではない）**: 管理対象ファイルの隣に
  `.kizami` ファイルを1つずつ置く。プロジェクト全体を1つの集約ファイルで管理する案も検討したが、
  ファイルが肥大化すること・複数人が同時に追記するとマージコンフリクトが発生すること・
  kizamiがMarkdown ADRで採用している「1つの関心事 = 1ドキュメント」の原則と一致しないことから却下した。

## Consequences

- 小さなサイドカーファイル1つで、任意の拡張子のファイルをkizamiで管理できるようになります。
- `kizami blame`・`kizami audit`・`kizami list`・`kizami show` がサイドカーファイルを自動的にサポートします。
- サイドカーフォーマットは意図的に最小限にしています。理由や背景はリンク先のMarkdown ADR/設計書に書きます。

## Alternatives Considered

- **CSVコメントメタデータ** (`# kizami:related: ...`): ファイルを汚す・フォーマット固有・フォーマットごとにパーサーが必要。
- **description フィールドを持つリッチなサイドカー**: 価値はあるが Markdown ドキュメントとの境界が曖昧になるため見送り。

## Related Files

- internal/decision/sidecar.go
- internal/decision/generate.go
- internal/decision/audit.go
- internal/search/blame.go
