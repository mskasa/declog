# コンテンツベースのドキュメント検出による任意ファイル名の受け入れ

- Date: 2026-04-14
- Status: Active
- Author: masahiro.kasatani

## Context

kizami はこれまで、`YYYY-MM-DD-slug.md` または レガシー形式の `NNNN-slug.md` にマッチするファイルのみを管理対象ドキュメントとして認識していた。これらのパターンに一致しないファイル — `ARCHITECTURE.md`、`API-SPEC.md`、チームが既に持つドキュメントなど — は `kizami list`・`kizami audit` およびその他のコマンドにサイレントに無視されていた。

このファイル名制約は、別のドキュメント管理システムから移行するチームにとって摩擦を生んでいた。既存のドキュメントをすべてリネームしなければ kizami で管理できないためである。

## Decision

以下の **両方** のマーカーを含む `.md` ファイルを kizami ドキュメントとして認識する：

1. `- Status:` で始まる行
2. `## Related Files` セクション見出し（ドリフト検出に必須）

両方のマーカーが存在する場合のみ対象となる。どちらか一方しか含まないファイルは kizami ドキュメントとして扱わない。

既存の命名規則（`YYYY-MM-DD-*.md`、`NNNN-*.md`）に一致するファイルは引き続きファイル名パターンマッチのみで認識される（高速パス、I/O なし）。規則外のファイルはコンテンツを読み取って判定する（低速パス）。

**任意ファイル名の slug**：`.md` 拡張子を除いたファイル名そのもの。
例：`ARCHITECTURE.md` → slug `ARCHITECTURE`

**`kizami list` のソート順**：フロントマターの `- Date:` フィールドを優先し、なければファイルの更新日時（mtime）を使用する。いずれも取得できない場合はリストの末尾に配置する。

## Consequences

- チームは kizami を段階的に導入できる。既存ドキュメントにマーカーを2つ追加するだけで管理対象となり、リネームは不要。
- `kizami list`・`kizami show`・`kizami blame`・`kizami audit`・`kizami search` がすべて任意ファイル名のドキュメントを透過的に扱う。
- パターン外ファイルのコンテンツスキャンは初回実行時に追加の I/O を発生させる。一般的な `documents.dirs` のサイズ（数十〜数百ファイル程度）では影響は無視できる。
- 2つのマーカーを必須とすることで、README や Changelog など kizami と無関係な `.md` ファイルが誤って管理対象にならないことを保証する。

## Alternatives Considered

**シングルマーカー認識（`- Status:` または `## Related Files` のいずれか）**
より緩やかだが、kizami 以外のドキュメントを誤って管理対象にするリスクがある。`## Related Files` セクションはドリフト検出の基盤であるため、これを必須とすることで `kizami audit` への参加を保証できる。

**オプトイン用フロントマターフラグ（例：`- kizami: true`）**
明示的だが、すべてのドキュメントに非標準フィールドを追加する必要がある。2マーカー方式は kizami 著者がすでに記述しているフィールドを再利用する。

**リネームを必須とする（現在の制約を維持）**
実装コストはゼロだが、既存チームの採用を不必要に困難にする。

## Related Files

- `internal/decision/generate.go`
- `internal/decision/decision_test.go`
