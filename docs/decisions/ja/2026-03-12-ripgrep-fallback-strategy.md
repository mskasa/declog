# ripgrepのフォールバック戦略

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

`kizami search`はMarkdownファイルに対して全文検索を実行する必要がある。
[ripgrep](https://github.com/BurntSushi/ripgrep)は代替手段より大幅に高速で`.gitignore`を尊重するが、外部バイナリであるため、すべてのマシンにインストールされているとは限らない。
ripgrepをハード依存にすると、インストールしていないユーザーで`kizami search`が動作しなくなる。

## Decision

`PATH`上にripgrep（`rg`）が存在する場合はそれを優先して使用し、存在しない場合はGoの標準ライブラリによる実装（`filepath.Walk` + `strings.Contains`）にフォールバックする。

## Consequences

- ripgrepがインストールされていないマシンでも`kizami search`が動作する
- ripgrepを持つユーザーはより高速な検索と`.gitignore`の考慮を享受できる
- フォールバック実装はシンプルだが、`docs/decisions/`ディレクトリの典型的な規模には十分
- ripgrepパスを実行するテストにはスキップ条件が必要（`rg`が見つからない場合は`t.Skip`）

## Alternatives Considered

- **ripgrepをハード依存にする：** コードがシンプルになるが、未インストールのユーザーで動作しない
- **Go標準ライブラリのみ：** 最大のポータビリティだが大規模リポジトリでは遅く、`.gitignore`の考慮がない
- **検索ライブラリの組み込み（`blevesearch`など）：** 高機能だが、シンプルな用途に対してバイナリサイズと複雑さが大幅に増加する

## Related Files

- `internal/search/search.go`
- `internal/search/search_test.go`
