# Related Filesのマッチングを単一の定義に統合する

- Date: 2026-08-03
- Type: ADR
- Status: Active
- Author: masahiro.kasatani

## Context

kizami は現在、「Related Filesのエントリが指定ファイルにマッチするか」を2箇所で独立に計算している。

- `search.Blame`（`internal/search/blame.go`）：ディレクトリエントリ（末尾スラッシュ規約）については `blameDirEntries` が独自に `strings.HasPrefix` チェックをインライン実装している。加えて `Blame` は、ドキュメント本文全体に対するファイルパスの全文検索も行う——これは構造的マッチとは異なる、より広い「言及」の概念である。
- `decision.CheckHook`（`internal/decision/hook.go`）：`hookPathMatches` が完全一致・ディレクトリ前置のチェックを独自にインライン実装している。

これらは単に「同じルールの2つの実装」ではない——実際に判定結果が食い違う。`blameDirEntries` は末尾が `/` のエントリだけをディレクトリとして扱う（`internal/db` のような素のエントリは完全にスキップする）のに対し、`hookPathMatches` は末尾スラッシュの有無を問わず*どのエントリも*完全一致とディレクトリ前置の両方でマッチさせる（`hook_test.go` の `TestCheckHook_DirectoryMatch` で、末尾スラッシュなしの `internal/db` エントリが `internal/db/db.go` にマッチすることが明示的にアサートされている）。あるドキュメントのRelated Filesエントリが、`kizami hook` では対象ファイルを縛ると判定される一方で、`kizami blame` の構造的チェックではカバーしていないと判定される、という状態が現状すでに起きうる。

両者とも「`file` はこのドキュメントのRelated Filesの1つか」という同じ問いに、別々の、しかも食い違う実装で答えている。`internal/context`（[[agent-context-layer-design]] 参照）はこれをエージェント向けAPI（`kizami context`、将来的には `kizami mcp`）として公開しようとしている。「このファイルを縛る決定は何か」という問いに、呼び出したコマンドによって答えが変わるAPIは成立しない。

利用者向けに公開されているドキュメント（`docs/site/adr-guide.md`：「ディレクトリも列挙できます——kizamiはその配下の全ファイルを関連ファイルとして扱います」）には、末尾スラッシュを要求する記述は一切ない。したがって `blameDirEntries` の厳密な挙動は、意図的に維持すべきドキュメント化された規約ではなく、むしろドキュメントに書かれた機能を下回っている。`hookPathMatches` の寛容な挙動こそが、ユーザーに説明されている機能と実際に一致するものである（テストレベルでも確認できる：`hook_test.go` の `TestCheckHook_DirectoryMatch` と `blame_test.go` の `TestBlameDirEntries_FileEntryIgnored` は、末尾スラッシュなしの `internal/db`／`database` エントリをネストしたファイルに対して評価した際、正反対の結果をアサートしている）。

一方、`decision.Audit`（`internal/decision/audit.go`）は関連するが別の問い——「Related Filesのエントリが今も実在するものを指しているか」——に、エントリごとの直接の `os.Stat` で答えている。これは（候補ファイルを受け取らない）マッチング操作ではないため同じ関数には統合しないが、globエントリが登場した後も正しく動作する必要がある：globの文字列（例：`internal/**/*_test.go`）に対する `os.Stat` は、その文字列そのままのファイルは存在しないため、決して成功しない。

また今回の設計では glob 対応（例：`internal/**/*_test.go`）が求められているが、現状の3つの呼び出し箇所のいずれにも存在しない。

## Decision

`internal/decision`（既に `ParseRelatedFiles` を持つパッケージ）に小さな共有プリミティブを2つ追加し、既存の全呼び出し箇所をこちらに寄せる。

**`Match(entry, file string) (kind MatchKind, ok bool)`** — 「Related Filesのエントリが候補ファイルにマッチするか」の唯一の定義。Contextで述べた食い違いを解消するため、より緩やかな `hookPathMatches` 側のルールを正とする（`blameDirEntries` の厳密なルールがマッチしていたものは全て、こちらでもマッチする上位互換になっている）。
1. `*` または `?` を含む → **glob**。`/` 区切りのセグメント単位で `path.Match` のセマンティクスを適用し、加えて `**` セグメントが0個以上のパスセグメントにマッチする拡張を明示的にサポートする（Go標準の `path.Match` には `**` がないため、依存を追加せず約20行の再帰的セグメントマッチャーとして実装する）。
2. それ以外の場合、末尾の `/` は取り除く（見た目上の違いのみで判定には影響しない）。`file == entry` なら **exact**、`file` が（スラッシュを除いた）エントリをパス構成要素の前置として持つなら **dir** としてマッチする——エントリ自体に末尾スラッシュが付いているかどうかは問わない（`internal/db` と `internal/db/` はRelated Filesのエントリとして全く同じに振る舞う）。

**`EntryExists(repoRoot, entry string) bool`** — 「このエントリが今も実在するものを指しているか」の唯一の定義。完全一致・ディレクトリエントリについては `os.Stat` に委譲する（挙動は現状のまま）。globエントリについては未チェックのまま `true` を返す——globが実際に少なくとも1つの実ファイルにマッチしているかを検証するにはディレクトリ走査が必要であり、本ステップのスコープ外とする（[[agent-context-layer-design]] のOpen Questionsで追跡する）。

更新する呼び出し箇所：
- `decision.CheckHook` — `hookPathMatches` を削除し、`Match` を呼ぶようにする。既存の（globでない）エントリの挙動は変わらない。副次的に、pre-commitフックでもRelated Filesのglobエントリが利用可能になる。
- `search.blameDirEntries` — インラインのディレクトリチェック（従来は末尾スラッシュ必須）を `Match` に置き換え、`dir`/`glob` の種別のみを対象とする（`exact` 種別は意図的に除外する。完全一致エントリは既に `Blame` の全文検索で見つかるため、ここでも含めると単なる重複作業になるだけで、後段で重複除去されるとはいえ無駄である）。本ADRで唯一の意図的な挙動変更はここである：`kizami blame` が、末尾スラッシュなしの素のディレクトリ風エントリ（例：`internal/db`）を、`kizami hook` が既にそうしていたのと同じように構造的にマッチできるようになり、Contextで述べたギャップが解消される。本リポジトリの現行ドキュメント（`docs/decisions/`・`docs/design/`）を確認した限り、末尾スラッシュなしのディレクトリエントリを使っているものは無いため、現時点で観測可能な影響はない。
- `decision.Audit` — インラインの `os.Stat` 呼び出しを `EntryExists` に置き換える。現状のエントリ（完全一致・ディレクトリ）に対する挙動は変わらない。globエントリが登場した際には、常に「missing」と誤報告されるのではなく、正しくスキップされるようになる。
- `internal/context.Resolve`（新規）— 主要な新規利用者。`Match` でクエリされたファイル群を縛る決定を特定し、`EntryExists` でマッチした各決定のドリフト状態を計算する。

## Consequences

- 「エントリXがファイルYにマッチするか」の実装が1つになり、`kizami hook`・`kizami blame` の構造的な半分・`kizami audit` の存在チェック・新規の `kizami context` すべてで再利用される。マッチングのバグ修正やエントリ構文の拡張が1箇所で完結する。
- `## Related Files` のglobエントリは、`kizami audit` のドリフトチェック（明示的にスキップされる。設計ドキュメントのOpen Questions参照）を除き、全コマンドで意味を持つようになる。一部コマンドでは動くが他では黙って機能しない、という状態を避けられる。
- `kizami hook` の挙動は、現在パスしているテスト（`hook_test.go` が本ADRの維持すべき仕様そのもの）に関して変わらない。
- `kizami blame` の構造的ディレクトリエントリマッチングは、末尾スラッシュなしのディレクトリ風エントリも `kizami hook` が既にそうしていた通り、また `docs/site/adr-guide.md` が既にドキュメント化していた通りに認識するようになる（Decision参照）。旧来の（ドキュメント化されていない挙動を前提とした）`TestBlameDirEntries_FileEntryIgnored` は `TestBlameDirEntries_BareDirEntryMatches` にリネームし、マッチする方をアサートするよう更新する。全文「言及」検索はいずれにせよ構造的マッチングとは別のまま残る。
- `kizami audit` の既存の完全一致・ディレクトリエントリのチェックは変わらない。インラインの `os.Stat` ロジックを重複させなくなるだけである。

## Alternatives Considered

**`Blame` の全文検索も `Match` に統合する（「言及」と「関連」の区別をなくす）**
`kizami blame` のコードパスは1本になるが、「このドキュメントが正式にファイルにリンクしている」（ドリフトチェック対象、エージェントが権威あるものとして扱ってよい）と「このドキュメントがたまたまその文字列を含んでいる」（ドリフトチェック対象外、例示など偶然の一致もしばしば）を混同してしまう。`kizami context` にはエージェント向けの強いシグナルが必要であり、コード上で概念を分離しておくことでそれを保てる。`kizami blame` は人間向けの「関連しうる全て」の検索として、別の正当な役割を持ち続ける。

**このステップで `os.Stat` ベースの存在チェックをglobエントリにも対応させる（ディレクトリ走査による）**
長期的には正しい挙動だが、（パターンに対する木全体での）*いずれかの*マッチの存在確認は、本ADRがスコープとするファイルマッチングとは本質的に別の操作である。また、実際のドキュメントにglobエントリが登場するまでは不要でもある。見送り、設計ドキュメントのOpen Questionsで追跡する（ここで先回りして解決しない）。

**サードパーティのglobライブラリ（例：`doublestar`）を使う**
`**` をより堅牢に扱え、より多くのglob構文をカバーできるが、実際に必要な部分集合（セグメント単位の `*`/`?` に加えて `**` ワイルドカードセグメント）は約20行で直接実装できるほど小さい。CLAUDE.md の方針は、標準ライブラリに近い労力で済む場合は依存を避けることを志向している（`use-go-over-shell-script` 参照）。

## Related Files

- `internal/decision/match.go`（新規）
- `internal/decision/audit.go`
- `internal/decision/hook.go`
- `internal/search/blame.go`
- `internal/context/`（新規）
