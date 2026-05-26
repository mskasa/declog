# kizami lint における空の Related Files をエラーではなく警告として扱う

- Date: 2026-05-26
- Status: Active
- Author: masahiro.kasatani

## Context

`kizami lint` は当初、ドキュメントの `## Related Files` セクションが空の場合にエラーを報告していた。
実運用では、Related Files を空にしておくことが正当なケースが存在する。

- ドキュメントが実装よりも先に書かれる場合（設計段階で作成する ADR など）
- リポジトリ全体に関わるアーキテクチャ・プロセス系のドキュメントで、特定のファイルへの紐付けが馴染まない場合

このような状況でエラーとして CI をブロックしても、ドキュメントの品質向上にはつながらない。
また `kizami audit` および `kizami blame` は Related Files が空のドキュメントをすでにスキップしており、
ツールのレベルでは空の状態がすでに正常なケースとして扱われていた。

## Decision

`kizami lint` は Markdown ドキュメントの `## Related Files` セクションが存在しない・空の場合、
`[error]` ではなく `[warn]` を出力するよう変更した。
警告のみの場合は exit code 0 で終了するため、CI はブロックされない。

sidecar（`.kizami`）ファイルはエラーのまま据え置く。
sidecar は他のファイルに紐付けることを唯一の目的として存在するため、
`related:` リストが空であることは常に構造的な欠陥である。

## Consequences

- Related Files が空のドキュメントは CI を通過するため、ドキュメント作成初期段階の摩擦が減る。
- `kizami audit` および `kizami blame` の動作に変更はない（すでにスキップしていた）。
- 警告をエラーとして扱いたいチームは、シェルレベルで対応が必要になる場合がある。
- 実装完了後も Related Files が空のまま放置されるリスクがある。
  これはプロセス上の課題であり、警告によって継続的なリマインダーとして機能する。

## Alternatives Considered

- **Draft ステータスのみスキップ** — より的を絞った対応だが、Related Files が本質的に存在しないアーキテクチャ系ドキュメントには対応できない。
- **`--strict` フラグの追加** — 呼び出し側に判断を委ねる形だが、明確なデフォルトを持たない複雑さを生む。
- **チェックを完全に削除** — リマインダーとしての価値も失われるため不採用。警告方式を選択した。

## Related Files

- `internal/decision/lint.go`
- `internal/decision/lint_test.go`
- `cmd/lint.go`
