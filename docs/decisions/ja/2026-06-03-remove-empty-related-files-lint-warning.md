# kizami lint の Related Files 空警告を削除する

- Date: 2026-06-03
- Status: Active
- Author: masahiro.kasatani
- Supersedes: lint-empty-related-files-as-warning

## Context

`kizami lint` はこれまで、Markdown ドキュメントの `## Related Files` セクションが空の場合に
`[warn]` を出力していた。これは記入漏れへの注意喚起が目的だった。

実際の運用では、`## Related Files` が空であることが「作業途中」ではなく「正しい最終状態」
であるケースが複数存在することが判明した：

- **機能削除 ADR**: 関連するソースファイルが同一コミット内で意図的に削除されるため、
  記載できるファイルが存在しない。
- コードベース全体にわたるアーキテクチャ文書やプロセス文書。
- 実装前に書くドラフト段階のドキュメント。

正当な空ケースが増えるにつれ、警告は「記入漏れかもしれない」というシグナルとしての
精度を失い、ノイズになりかねない。

また、`kizami hook pre-commit` がコミット時のリマインダーとして機能しており、
「Related Files を書き忘れた」ケースはすでに hook によってカバーされている。

## Decision

`kizami lint` の Markdown ドキュメントにおける `## Related Files` 空警告を削除する。

空のセクションは警告なしに受け入れる。`## Related Files` セクション自体はドリフト検出
（`kizami audit`、`kizami blame`、`kizami hook`）に必要なため引き続き必須だが、
空の場合は lint からは無視される。

サイドカー（`.kizami`）ファイルはエラー動作を維持する。サイドカーは他ファイルに
アノテーションを付けるためだけに存在するため、`related:` リストが空であることは
常に構造的な欠陥である。

## Consequences

- `kizami lint` は `## Related Files` が空のドキュメントに対して何も報告しなくなる。
- 機能削除 ADR など「意図的に空」なドキュメントが lint を正常に通過するようになる。
- Related Files の記入リマインダーとしては、pre-commit hook が主要な手段として残る。
- 意図せず空のまま放置されるリスクはわずかに上昇するが、hook でカバーされているため許容できる。

## Alternatives Considered

**警告を抑制する明示的マーカーを追加する（例: `<!-- intentionally empty -->`）**
意図的な空と記入漏れを区別できるが、著者に新たなルールの習得・適用コストを課す。
得られるメリットに対してオーバーヘッドが大きい。

**警告を維持し、正当な空ケースをドキュメント化する**
シグナルは保たれるが、既知の空ケースでノイズを許容し続けることになり、
正当なケースが増えるほどシグナルの精度はさらに低下する。

## Related Files

- internal/decision/lint.go
- internal/decision/lint_test.go
