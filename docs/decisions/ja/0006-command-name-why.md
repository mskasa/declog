# 0006: コマンド名を「why」にする

- Date: 2026-03-12
- Status: Superseded by 0010
- Author: masahiro.kasatani

## Context

CLIツールには短く・覚えやすく・目的を伝えるコマンド名が必要だった。
リポジトリ名は`declog`（decision log）だが、実行コマンド名は別の問題として検討する。
開発中に頻繁に使うコマンドなので、簡潔さは重要な要素だった。

## Decision

CLIコマンド名を`dec`・`declog`・`adr`ではなく`why`にする。
「why」はこのツールの目的——意思決定の理由を記録すること——を直接表現している。

## Consequences

- コマンドが文脈のなかで自然に読める：`why log "use postgres"`・`why list`・`why show 3`
- 頻繁に打つのに十分な短さ
- 覚えやすく自己説明的——新しいチームメンバーもツールの目的をすぐに理解できる
- `why`は珍しいコマンド名なので、既存のシステムコマンドとの衝突はほぼ起きない
- バイナリは`why`として配布するが、検索やパッケージ管理での曖昧さを避けるためにリポジトリ名は`declog`のまま

## Alternatives Considered

- **`declog`：** リポジトリ名と一致するが長く、動詞として表現力が低い
- **`dec`：** 短いが意味が曖昧（decimal? declare?）
- **`adr`：** 成果物の種類を表すが行動を表さない。また既存の`adr-tools`プロジェクトと衝突する
- **`record`：** 説明的だが汎用的すぎて他のツールと衝突しやすい

## Related Files

- `main.go`
- `.goreleaser.yaml`
