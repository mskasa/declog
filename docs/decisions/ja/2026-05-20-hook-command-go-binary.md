# pre-commitフックのロジックをGoバイナリに移行する（kizami hook pre-commit）

- Date: 2026-05-20
- Status: Active
- Author: masahiro.kasatani

## Context

従来のpre-commitフックは、シェルスクリプト（`internal/initializer/templates/pre-commit`）として実装されており、`docs/decisions/` 配下のファイルがMarkdown以外のファイルと一緒にステージされているかを確認し、されていなければADR作成を促す仕組みだった。

このアプローチには3つの問題があった：

1. **設定非対応**: スクリプトが `docs/decisions/` をハードコードしており、`kizami.toml` の設定（複数ドキュメントディレクトリ、カスタムパスなど）を参照できなかった。
2. **Related Filesチェックなし**: 新しいドキュメントの作成を促すのみで、ステージされたファイルが既存ドキュメントのRelated Filesに含まれているかを検出できなかった。
3. **Windows非対応**: シェルスクリプトはWSLやGit BashなしにはWindowsでCI動作せず、kizamiのクロスプラットフォーム対応という方針と矛盾していた。

## Decision

pre-commitフックのロジック全体をGoのサブコマンド `kizami hook pre-commit` に移行する。

シェルスクリプトのテンプレートは、薄いラッパーに変更する：

```sh
#!/bin/sh
if command -v kizami >/dev/null 2>&1; then
  kizami hook pre-commit
fi
```

Goコマンドは2つのチェックを行う：

1. **Related Filesチェック**: 設定されたすべてのドキュメントディレクトリのActiveドキュメントを走査する。ステージされたファイルが、ステージされていないドキュメントのRelated Filesに含まれている場合、そのドキュメントの確認・更新を促すメッセージを表示する。
2. **新規ドキュメントチェック**: ドキュメントがステージされておらず、Markdown以外のファイルが含まれている場合に、新しいドキュメント作成を促すメッセージを表示する。これは従来のシェルスクリプトと同じ動作だが、設定に対応している。

どちらのチェックも通知のみ。コマンドは常に終了コード0を返し、コミットをブロックしない。

## Consequences

- pre-commitの動作が設定対応になる：`kizami.toml` の `audit.dirs` / `documents.dirs` を参照する。
- Related Filesチェックにより、新しいドキュメントを作成すべき場面だけでなく、既存の意思決定を更新すべき場面でも通知が行われるようになる。
- WindowsユーザーもUnixユーザーと同じフック動作を得られる。
- `kizami init` 後にkizamiがアンインストールされた場合、フックが機能しなくなる。これは許容できるトレードオフ：フックはkizamiが存在することを前提として生成されている。`kizami` が見つからない場合は何もせずに終了するためコミットは妨げない。

## Alternatives Considered

**シェルスクリプトの拡張**: `grep` とパス操作で実装は可能だが、シェルでTOMLを読み込むことは現実的でなく、設定対応が困難。

**専用バイナリ（kizami-hook）**: `command -v kizami` への依存を回避できるが、配布の複雑さが増す。既存バイナリのサブコマンドで十分。

## Related Files

- internal/initializer/templates/pre-commit
- internal/initializer/hook.go
- internal/decision/hook.go
- cmd/hook.go
