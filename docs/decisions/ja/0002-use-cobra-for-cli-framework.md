# 0002: CLIフレームワークにCobraを使用する

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

declogは複数のサブコマンド（`log`・`list`・`search`・`show`・`status`）を提供する。
サブコマンドのルーティング・フラグ解析・ヘルプテキスト生成をゼロから実装するのは、反復作業が多くエラーが生じやすい。
これらの関心事を一貫して処理するためのCLIフレームワークが必要だった。

## Decision

CLIフレームワークとして[cobra](https://github.com/spf13/cobra)を使用する。
CobraはGoのCLIアプリケーションにおけるデファクトスタンダードであり、サブコマンド管理・自動ヘルプ生成・シェル補完をすぐに利用できる。

## Consequences

- サブコマンド構造とフラグ解析がボイラープレートなしで一貫して処理される
- Bash・Zsh・Fish・PowerShellのシェル補完が最小限の実装で利用可能
- Goエコシステムで広く使われており（Kubernetes・Hugo・GitHub CLIなど）、コントリビューターにとって馴染みやすい
- 外部依存関係が増えるが、成熟した安定したライブラリである

## Alternatives Considered

- **`flag`（標準ライブラリ）：** サブコマンドのサポートがなく、複数コマンドのCLIには低レベルすぎる
- **`urfave/cli`：** 有力な選択肢だが、GoエコシステムではCobraほど広く普及していない
- **手動ルーティング：** `os.Args`に対するシンプルな`switch`文——少数のコマンドには十分だが、スケールしないうえにヘルプ/補完の生成機能がない

## Related Files

- `cmd/root.go`
- `cmd/log.go`
- `cmd/list.go`
- `cmd/search.go`
- `cmd/show.go`
- `cmd/status.go`
