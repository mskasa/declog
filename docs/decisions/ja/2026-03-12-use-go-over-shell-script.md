# シェルスクリプトではなくGoを使用する

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

declogはLinux・macOS・Windowsで動作するツールとして配布する必要がある。
シェルスクリプトはプラットフォーム間での移植性がなく、Bashは追加のセットアップなしにWindowsで利用できない。
また、ファイルI/O・Git連携・テキスト処理を安定して扱う実装言語が必要だった。

## Decision

シェルスクリプトではなく、Goでdeclogを実装する。
Goは単一の静的バイナリにコンパイルでき、すべての対象プラットフォームをサポートし、型安全性と豊富な標準ライブラリを提供する。

## Consequences

- GoReleaserを使った単一バイナリ配布が可能——ランタイムの依存関係が不要
- WSLやBashなしでWindowsをサポートできる
- 型安全性により、シェルスクリプトで起きがちな実行時エラーのクラスを排除できる
- シェルスクリプトと比べてコントリビューションの敷居はやや高くなるが、GoエコシステムにおけるCLIツールの標準的な選択肢である

## Alternatives Considered

- **シェルスクリプト（Bash）：** 記述が簡単だがWindowsへの移植性がなく、テストが難しい
- **Python：** クロスプラットフォームだが、ユーザーのマシンにPythonランタイムが必要で、パッケージングが複雑
- **Node.js：** クロスプラットフォームだが、Node.jsランタイムが必要で、バイナリ配布にはバンドルツール（pkgなど）が必要

## Related Files

- `go.mod`
- `main.go`
