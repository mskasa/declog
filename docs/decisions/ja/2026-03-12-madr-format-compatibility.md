# MADRフォーマットとの互換性

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

ADR（アーキテクチャ上の意思決定記録）ツールはさまざまなMarkdownフォーマットを使用する。
独自のフォーマットを選択すると、他のツールとの移行・連携が難しくなる。
すでにADRツールを使用しているチームがdeclogを採用する際の摩擦を最小限に抑えたい。

## Decision

[MADR（Markdown Architectural Decision Records）](https://adr.github.io/madr/)と互換性のあるテンプレートを採用する。
テンプレートはContext・Decision・Consequences・Alternatives Consideredのセクションで構成される。

## Consequences

- declogで書かれた意思決定記録は、他のMADR互換ツールで読み取れる
- フォーマットは人間が読みやすく、特別なツールなしで閲覧できる
- 一貫した構造により、プロジェクト内の意思決定を素早くスキャン・比較しやすい
- テンプレートは意図的に最小限に抑えており、チームは必要に応じてセクションを拡張できる

## Alternatives Considered

- **Nygardフォーマット（ADRの原型）：** よりシンプル（Context・Decision・Status・Consequences）だが表現力が低く、Alternativesセクションがない
- **独自フォーマット：** 最大の柔軟性を持つが、既存ツールとの互換性がない
- **YAMLやTOMLのフロントマター：** 機械可読なメタデータを付与できるが、インラインでの可読性が下がり、パースの複雑さが増す

## Related Files

- `internal/template/template.go`
