# slug検索を、ネストした言語バリアントに対しても再帰的にする

- Date: 2026-08-03
- Type: ADR
- Status: Active
- Author: masahiro.kasatani

## Context

`decision.FindBySlug(dir, slug)` は `dir` を再帰的に走査するが、最初のマッチで止まる（`filepath.SkipAll`）。本リポジトリ自身のEN/JAドキュメントペアは、別々のトップレベル設定ディレクトリではなく、*同じ*設定ディレクトリの配下にネストしている（例：`docs/decisions` の下の `docs/decisions/ja/`）。`filepath.WalkDir` はエントリを辞書順で訪問するため、EN側のファイル（`2026-...-slug.md`）が常に先に見つかり、`ja/` に降りる前に走査が止まる——JA側の対応するファイルは一度も見られることさえない。

これは `kizami mcp` の `kizami_get_decision` ツール（[[mcp-tools-as-questions-not-verbs]] 参照）を実装する過程で発見された。あるslugにマッチする全ドキュメントを返す必要があったが、`FindBySlug` を再利用するとネストしたバリアントが黙って落とされることが分かった。同じ基盤関数は、`cmd.findAllBySlug`／`cmd.findBySlug` 経由で既存の3つのコマンドを支えている。

- `kizami show <slug>`（複数マッチを許容し、全部表示する）——今日では、ネストしたペアのEN側しか表示しない。両方は決して表示されない。
- `kizami status <slug> <status>` と `kizami supersede <slug> ...`（特定のファイルを変更しようとしているため、曖昧な場合はエラーにする設計）——今日では、JA側の文書が存在することにも気づかずEN側の文書*だけ*を黙って変更してしまう。これは機能の欠落より悪い：データ整合性のリスクである。EN・JA両方のドキュメントを持つslugに対して `kizami status use-x active` を実行すると、JA側のStatusフィールドがEN側と黙って同期しなくなり、エラーも警告も出ない。

## Decision

`decision.FindAllBySlug(dir, slug) ([]*Decision, error)` を追加する：`FindBySlug` と同じ再帰的な走査だが、早期終了なしに全マッチを収集する。`FindBySlug` 自体は変更せず、既存の挙動と呼び出し元をそのまま維持する。

`cmd.findAllBySlug`（`show`／`status`／`supersede` を支える共有ヘルパー）だけを新しい関数を呼ぶように切り替え、各設定済みトップレベルディレクトリをループしながら、各ディレクトリ内の*全*マッチ（最初の1件だけではなく）を正しく収集するようにする。その曖昧性検出の利用元である `cmd.findBySlug` も、これによりネストした言語バリアントの曖昧性を正しく検出するようになる——これはクロスディレクトリの衝突に対して既に設計されていた挙動そのものであり、単にこのケースが見えていなかっただけである。

`cmd/log.go` の2箇所の `decision.FindBySlug` への直接呼び出し（新規ADR作成時の `--supersedes <slug>` の解決）は、意図的に変更しない。このパスも曖昧性検出にしてしまうと、EN/JAペアを持つslugに対する `--supersedes` が完全にブロックされてしまい、現在のCLI表面には（パスではなくslugしか受け取らないため）それ以上曖昧性を解消する手段がない——これは単なる厳格化ではなく、実質的な機能の後退になる。`--supersedes` に同様の対応が必要になった場合の、別途のフォローアップとして残す。

## Consequences

- `kizami show <slug>` は、ネストした言語バリアントを含め、あるslugにマッチする全ドキュメントを正しく表示するようになった——バイリンガルペアの半分がコマンドから見えなくなっていた実在のギャップを閉じる。
- `kizami status`／`kizami supersede` は、ネストしたバリアントによってslugが曖昧な場合、正しく処理を*拒否*するようになった。クロスディレクトリの衝突に対して既に表示していた「どのファイルか指定してください」というエラーと同じものが表示される。これは挙動の変更である：以前は黙って成功していた（EN側だけを変更していた）コマンドが、今はエラーになる。これは意図した修正である——バイリンガルペアの片方だけを黙って変更することこそが実際のバグだった——が、この古い黙った挙動に依存していたワークフローがあれば調整が必要になる（現時点ではこれらのコマンドにパスによる曖昧性解消の手段がない。実運用で支障が出るようであれば、将来的に用意する必要がある）。
- `kizami adr ... --supersedes <slug>` は、ネストしたバリアントを持つslugに対して現状の挙動（最初のマッチが勝つ）を維持する——ここでは修正せず、フォローアップとして追跡する。

## Alternatives Considered

**新しい関数を追加する代わりに、`FindBySlug` 自体を修正する（早期終了を取り除く）**
`cmd/log.go` の `--supersedes` パスも修正できるが、`FindBySlug` は単一の `*Decision` を返す。曖昧なマッチは、黙って1つを選ぶ（まさに修正対象のバグ）か、エラーにするかのいずれかになる。後者は `--supersedes` にとって、EN/JAペアを持つslugをsupersedeする唯一の手段を、代替手段なしに失わせることになる。戻り値の型を変更し、全呼び出し元にその処理を強制するのは、今日実際に壊れていることが分かっている1つのパス（`show`／`status`／`supersede`）を修正するより、大きくリスクの高い変更である。

**バグを放置し、`kizami_get_decision` の新規コードパスだけで対応する**
直近の必要は満たせるが、`kizami status`／`kizami supersede` の「黙った部分変更」リスクはそのまま残る——これはMCP関連のものとは無関係な正当性の問題であり、Agent Context Layerのロードマップとは無関係に、単独で修正する価値がある。

## Related Files

- `internal/decision/generate.go`
- `cmd/root.go`
