# kizami — CLAUDE.md（日本語版）

## プロジェクト概要

ドキュメントとコードの乖離を防ぐ、Go製のリビングドキュメント管理CLIツール（`kizami`）。
ドキュメントは `docs/decisions/` 配下（設定可能）に Markdown で保存し、Git で管理する。

**コアバリュー：** Markdown ドキュメントの `## Related Files` セクションがソースファイルとドキュメントを結ぶ唯一の接点。
`kizami audit` はそのファイルが削除・移動されていないかを検証し、ドキュメントの陳腐化を防ぐ。

もともと ADR（アーキテクチャ決定記録）専用ツールとして開発されたが、設計書・API仕様書・
アーキテクチャ概要など、あらゆるリビングドキュメントに対応するよう拡張中。

---

## ディレクトリ構成

```
kizami/
├── cmd/
│   ├── root.go         # ルートコマンド（kizami）
│   ├── log.go          # kizami adr / kizami design
│   ├── list.go         # kizami list
│   ├── search.go       # kizami search
│   ├── show.go         # kizami show
│   └── status.go       # kizami status
├── internal/
│   ├── decision/
│   │   ├── decision.go     # Decision型の定義・パース
│   │   ├── generate.go     # ファイル生成・自動採番ロジック
│   │   └── decision_test.go
│   ├── search/
│   │   ├── search.go       # キーワード検索
│   │   └── search_test.go
│   └── template/
│       └── template.go     # Markdownテンプレート管理
├── docs/
│   └── decisions/          # このリポジトリ自身のADR（ドッグフーディング）
│       ├── 2026-03-12-use-go-over-shell-script.md
│       ├── 2026-03-12-use-cobra-for-cli-framework.md
│       ├── 2026-03-12-madr-format-compatibility.md
│       ├── 2026-03-12-plaintext-markdown-only.md
│       └── 2026-03-12-ripgrep-fallback-strategy.md
├── CLAUDE.md
├── CLAUDE.ja.md        # 日本語版（本ファイル）
├── go.mod              # module github.com/mskasa/kizami
├── go.sum
└── main.go
```

---

## 技術スタック

| 用途              | ライブラリ／ツール                      | 選定理由                                         |
| ----------------- | --------------------------------------- | ------------------------------------------------ |
| CLIフレームワーク | [cobra](https://github.com/spf13/cobra) | Go CLIのデファクトスタンダード                   |
| テスト            | 標準 `go test`                          | 外部依存を増やさない                             |
| 検索              | ripgrep（外部コマンド）＋フォールバック | 高速検索。未インストール時は標準ライブラリで代替 |
| 配布              | GoReleaser + GitHub Actions             | シングルバイナリ配布                             |

- Goバージョン：1.22以上
- 対応OS：Linux / macOS / Windows（シングルバイナリ前提）

---

## コマンド仕様（MVP）

```bash
kizami adr "<title>"              # ADRを生成してエディタを開く
kizami design "<title>"           # 設計書を生成してエディタを開く
kizami list                       # 新しい順に一覧表示（スラッグ・日付・ステータス・タイトル）
kizami search <keyword>           # キーワード検索
kizami show <slug>                # 指定スラッグのDocumentを表示（例: kizami show use-go-over-shell-script）
kizami status <slug> <status>     # ステータス変更（例: kizami status use-postgresql superseded --by use-cockroachdb）
kizami blame <file>               # 指定ファイルに関連するDecisionを逆引き
```

### ステータス定義

| ステータス           | 意味                               | 使うタイミング                             |
| -------------------- | ---------------------------------- | ------------------------------------------ |
| `Active`             | 現在有効な判断（デフォルト）       | コード変更と同時にコミット                 |
| `Inactive`           | 単純に無効になった                 | 置き換え先のADRが存在しない場合            |
| `Superseded by <slug>` | 別のDocumentに置き換えられた     | 新しいDocumentを作成した場合               |

**ステータス運用方針：**
- デフォルトは `Active`。ADRはコード変更と同時にコミットする運用のため、作成時点で意思決定済みとみなす
- 設計を覆す新しいADRを作成した場合は既存ADRを `Superseded by <slug>` にする
- 置き換え先のADRが存在しない場合は `Inactive` にする

---

## Markdownテンプレート（MADR準拠）

`kizami adr` 実行時に生成されるテンプレート：

```markdown
# {Title}

- Date: {YYYY-MM-DD}
- Status: Active
- Author: {git config user.name}

## Context

<!-- なぜこの判断が必要になったか。背景・制約・問題を記述する -->

## Decision

<!-- 何を決めたか。1〜3文で明確に -->

## Consequences

<!-- この判断による影響・メリット・トレードオフ -->

## Alternatives Considered

<!-- 検討したが採用しなかった選択肢とその理由（省略可） -->

## Related Files

<!-- このDecisionに関連するファイルを列挙する（例: internal/search/search.go）。 -->
```

### ファイル命名規則

```
YYYY-MM-DD-kebab-case-title.md
例: 2026-03-12-use-go-over-shell-script.md
```

- `YYYY-MM-DD`：作成日（時系列ソートの基準）
- kebab-case：タイトルを小文字・ハイフン区切りに自動変換
- 保存先：`docs/decisions/`（リポジトリルートからの相対パス）
- このリポジトリのドッグフーディング用ADRは、英語版と日本語版の両方を作成する：
  - 英語版：`docs/decisions/2026-03-12-use-go-over-shell-script.md`
  - 日本語版：`docs/decisions/ja/2026-03-12-use-go-over-shell-script.md`

---

## 🐕 ドッグフーディング方針（最重要）

**このリポジトリ自体でkizamiを使って設計判断を記録する。**

### なぜドッグフーディングが重要か

- READMEの最強の説得材料になる（「作者自身が使っている」という事実）
- 書きづらいと感じた箇所がそのままUX改善のフィードバックになる
- GitHubを訪れた開発者が `docs/decisions/` を見るだけでツールの価値を理解できる

### Claudeへの指示

**実装中に以下のような判断が発生したら、必ずADRの作成を提案すること：**

- 技術選定（ライブラリ・アルゴリズム・ファイル形式）
- 複数の実装方針で迷った場合
- 既存の設計を変更・廃止する場合
- 将来の拡張に影響する設計上の決定

**ADR作成のトリガー例：**

```
「cobraを選んだ理由をADRに残しましょうか？」
「ripgrepのフォールバック戦略についてDecisionを記録します」
「この設計判断はdocs/decisions/に残す価値があります」
```

### ADRの粒度ガイドライン

**ADRに記録すべき判断：**

- 複数ファイル／複数コンポーネントに影響する設計判断
- 外部要因（負荷試験・障害対応・パフォーマンス計測など）を伴う判断
- 将来の開発者が「なぜこうなっているか」を知りたくなる判断

**ADRに記録しなくてよい判断：**

- 変数名・関数名などの小粒な変更
- 自明な実装詳細
- 1ファイル内で完結する理由（コードコメントに書く）

**コードコメントとの使い分け：**

| スコープ | 記録する場所 |
| -------- | ------------ |
| 1ファイル内で完結する理由 | コードコメント |
| 複数ファイルにまたがる理由 | ADR |
| 両方に該当する | 両方に書き、コメントにADRへのリンクを残す |

例 — コードコメントからADRへのリンク：

```go
// AuthorFromGit reads the author name from git config.
// Decision to use git config instead of an environment variable: docs/decisions/2026-03-16-allow-direct-adr-updates-with-git-history.md
func AuthorFromGit() string {
    ...
}
```

### ADRの更新原則

**ADRはGitで履歴管理されるため、直接更新して構わない。**
**変更履歴はgit logで追跡できる。**

**許容される操作：**
- 同じ判断の内容が変わった場合はADRを直接更新する
  → git diffで何が変わったか、git logでなぜ変えたかが追跡できる
- StatusをActive → Inactive または Superseded by <slug> に変更する
- 誤字脱字の修正
- Related Filesへの追記

**Supersededを使うケース：**
- 判断の方向性ごと変わった場合は新しいADRを作成してSupersededにする
- 同じ判断の修正・更新であれば直接更新で構わない

**ADR更新時のコミットメッセージ：**
- 何をなぜ変えたかを明記する
- 例：`docs: update ADR madr-format-compatibility - increase pool size from 10 to 20 based on load test`
- 悪い例：`update adr`

### 開発開始時点で作成すべき初期ADR

コードを1行も書く前に、以下のADRを手動で作成しておくこと：

| スラッグ                    | 内容                                                          |
| --------------------------- | ------------------------------------------------------------- |
| use-go-over-shell-script    | Goを選んだ理由（シングルバイナリ、Windows対応、型安全）       |
| use-cobra-for-cli-framework | cobraを選んだ理由（デファクト、シェル補完、サブコマンド管理） |
| madr-format-compatibility   | MADRフォーマット準拠の理由（既存ADRツールとの互換性）         |
| plaintext-markdown-only     | DBを使わずMarkdownのみにした理由（Git親和性、可搬性）         |
| ripgrep-fallback-strategy   | ripgrep依存とフォールバック設計の判断                         |
| command-name-why            | コマンド名をもともと `why` にした理由（rename-to-kizami-and-expand-scope で superseded） |

---

## 開発ルール

### コーディング規約

- コミット前に必ず `gofmt` / `goimports` を通す
- プッシュ前に `golangci-lint run` を実行し、指摘された問題をすべて修正する
- エラーはwrapする（`fmt.Errorf("...: %w", err)`）
- CLIの出力メッセージは**英語**に統一する
- コードコメントは**英語**に統一する

### GitHub Actions

- タグのすり替えによるインジェクションを防ぐため、アクションはバージョンタグではなくフルコミットSHAで指定する
- まずバージョンタグで記載し、その後 `pinact run` を実行してSHAに変換する：

```yaml
# 変換前
- uses: actions/checkout@v4

# 変換後（pinact run を実行）
- uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
```

#### アクションバージョンの更新

`pinact run --update` を実行すると、`.github/workflows/` 内の全アクションを最新バージョンに一括更新できる：

```bash
pinact run --update
```

> **注意：** `pinact` が走査するのは `.github/workflows/` のみ。
> `internal/initializer/templates/` に埋め込まれたワークフローテンプレートは自動更新の対象外。
> `.github/workflows/` のバージョンを変更したときは、テンプレート側も手動で合わせること。

### テスト方針

- 各パッケージに `_test.go` を置く
- ファイルI/Oを伴うテストは `t.TempDir()` を使う
- 外部コマンド（ripgrep等）に依存するテストはskip条件を入れる：

```go
if _, err := exec.LookPath("rg"); err != nil {
    t.Skip("ripgrep not installed")
}
```

### コミットメッセージ規約

```
<type>: <summary>

type:
  feat     新機能
  fix      バグ修正
  docs     ドキュメント（ADR追加も含む）
  refactor リファクタリング
  test     テスト追加・修正
  chore    ビルド・依存関係

例:
  feat: implement kizami adr command with auto-numbering
  docs: add ADR 0003 for MADR format compatibility
```

---

## ブランチ・PR運用

### ブランチ戦略

個人開発はシンプルに2種類のみ：

```
main
└── feature/xxx   # 機能単位で切る・完成したらmainへマージ
```

`develop` ブランチは作らない。個人開発では無意味に複雑になるだけ。

### ブランチ命名規則

```bash
feature/why-log-command
feature/why-list-command
feature/auto-numbering
docs/initial-adrs           # ADR追加もブランチを切る
fix/slug-generation-bug
```

### 1サイクルの流れ

GitのすべてはClaudeが実施する。オーナーは確認・承認・マージを担当する。

```
1. Claudeがfeatureブランチを作成する
2. Claudeが実装・コミット・プッシュする
3. ClaudeがPRを作成する（GitHub CLI: gh が必要）
4. オーナーがGitHub UIまたは `gh pr merge` でマージする
5. ClaudeがCLAUDE.mdの実装状況チェックボックスを更新する
```

mainブランチの履歴を綺麗に保つため、マージは常に **squash merge** を使う。
Claudeの作業中コミットは実装の詳細であり、1機能につき1コミットで十分。

### PRの説明テンプレート

PR作成時は必ず以下を記載する：

```markdown
## What
（変更内容を1文で）

## Why
（設計判断が伴う場合は関連ADRへのリンクを記載）

## Checklist
- [ ] テストが通ること（`go test ./...`）
- [ ] 設計判断があればADRを作成済みであること
- [ ] CLAUDE.mdの進捗を更新済みであること
```

---

## ClaudeとオーナーのRole分担

Claudeが実装を担当し、オーナーが判断・承認を担当する。

| 作業                              | 担当                   |
| --------------------------------- | ---------------------- |
| ブランチ作成                      | Claude                 |
| コード実装                        | Claude                 |
| テスト作成                        | Claude                 |
| コミット・プッシュ                | Claude                 |
| PR作成                            | Claude（`gh` CLI必要） |
| 設計判断が発生したらADR作成を提案 | Claude                 |
| **PRのレビュー・承認**            | **オーナー**           |
| **mainへのマージ**                | **オーナー**           |
| **ADRの内容確認**                 | **オーナー**           |
| **次に何をするかの意思決定**      | **オーナー**           |

### 理想的な会話の流れ

```
オーナー:
「CLAUDE.mdを読んで現在の状態を把握してください。
 feature/kizami-adr-commandブランチを作成して、
 kizami adrコマンドを実装してください。
 各ステップで確認を取りながら進めてください。」

Claude:
「CLAUDE.mdを確認しました。
 feature/kizami-adr-commandブランチを作成します。
 [ブランチ作成]
 kizami adrの実装を開始します...
 [実装]
 完了しました。自動採番のロジックで設計上の判断が発生しました。
 コミット前にADRを作成しますか？」

オーナー:
「はい、お願いします。」

Claude:
「docs/decisions/2026-03-23-auto-numbering-strategy.md を作成しました。
 コミット・プッシュします。
 PRを作成しますか？」

オーナー:
「はい。」

Claude:
「PRを作成しました: https://github.com/mskasa/kizami/pull/1
 レビューしてマージをお願いします。」
```

### Claudeへの行動原則

- **各主要ステップの前に確認を取る** — ブランチ作成→実装→コミット→PRを一気に進めない
- **設計判断が発生したらADRを能動的に提案する**
- **タスクを小さく保つ** — 1ブランチにつき1コマンド、1PRにつき1つの関心事
- **マージのたびにCLAUDE.mdの実装状況チェックボックスを更新する**

---

## ロードマップ

kizami はチームで実運用中。以下のロードマップはOSS公開に向けた目標と、実運用フィードバックから生まれる改善の両方を反映している。

### Phase 1 — 法的・信頼性の基盤整備 ✅

*より広いユーザーへ告知する前に必ず整えておくべき項目。*

- [x] `LICENSE` ファイルの追加（MIT）— ないと企業での利用が法務上ブロックされるケースが多い
- [x] `SECURITY.md` — 脆弱性の非公開報告手順を明記する
- [x] `CODE_OF_CONDUCT.md` — Contributor Covenant などの標準規約を採用する
- [x] `.github/ISSUE_TEMPLATE/` — バグ報告・機能要望のテンプレートを用意する
- [x] `.github/PULL_REQUEST_TEMPLATE.md` — 現在 CLAUDE.md 内にある PR テンプレートを正式化する
- [x] `CONTRIBUTING.md` の更新 — golangci-lint のバージョンとセットアップ手順が実態と乖離している

### Phase 2 — 品質と発見性の向上

*品質の底上げと、プロジェクトを見つけやすく・信頼されやすくする。*

**テストカバレッジ**
- [ ] チーム実運用で発見されたバグをリグレッションテストとして追加する

**CI**
- [x] テストマトリクスに Windows を追加
- [ ] テストマトリクスに macOS を追加（現時点では意図的にスキップ。Linux + Windows で十分と判断）

**GitHub リポジトリ整備**
- [x] Topics の設定（`cli`、`golang`、`documentation`、`adr`、`decision-record`、`living-documentation`）
- [x] README にバッジを追加（CI・Go Report Card・ライセンス・リリース）
- [x] README にチームでの活用事例や導入前後の比較を追記する

### Phase 3 — 配布とエコシステムの拡張

*より多様なチームに採用してもらえるよう、配布経路と拡張性を整える。*

**パッケージマネージャー**
- [ ] Homebrew formula
- [ ] Scoop（Windows ユーザー向け）

**GitHub Actions**
- [ ] GitHub Actions Marketplace 公開（`kizami audit`・`kizami lint` を再利用可能な Action として）

**拡張性**
- [ ] テンプレートのユーザー定義 — `kizami.toml` でテンプレートパスを指定可能にする

**ドキュメントサイト**
- [ ] チーム導入ガイド
- [ ] 移行ガイド（adr-tools・素の Markdown・Confluence/Notion からの移行）

### Phase 4 — Agent Context Layer（エージェント文脈層）

*kizami の重心を「人間が能動的にCLIを打つ」から「必要な瞬間に決定が自動で現れる」へ移す。AIエージェントは既に `docs/decisions/` を直接読めるため、到達性そのものは課題ではない。真の課題は、決定が下された瞬間に確実に書かれないこと、そして必要な瞬間に確実に提示されないことにある。*

**Step 1 — コンテキストリゾルバ**
- [ ] `internal/context` パッケージ：既存の2つの「関連ファイル」実装（`search.Blame` の全文検索と `decision.CheckHook` の構造化 `## Related Files` パース）を単一の定義に統合する
- [ ] Related Files のエントリに glob 記法（例：`internal/**/*_test.go`）を追加し、既存の完全一致・ディレクトリ前置マッチと併用可能にする
- [ ] `kizami context <files...> [--json] [--full]` — 変更ファイル群を渡すと、それらを縛るActiveな決定（および `supersededBy` を含むSuperseded決定）と、ファイルごとのドリフト状態を返す

**Step 2 — エージェント向けマニフェスト同期**
- [ ] `kizami agents sync` — CLAUDE.md / AGENTS.md 内にマーカー区間を維持し、どのパスがどの決定に縛られているかを一覧化する（ADR全文ではなくポインタの表）
- [ ] `kizami agents sync --check` — ADRのRelated Filesがマニフェストに反映されていない場合に失敗するCIチェック

**Step 3 — MCPサーバー**
- [ ] `kizami mcp` — リゾルバをMCPツールとして公開する。CLIの動詞をそのまま移植するのではなく、エージェントが問う「問い」の形にする：`kizami_decisions_for_files`、`kizami_search_decisions`、`kizami_get_decision`
- [ ] ツールの応答は既定で `## Decision` の要約のみを返す（`full` パラメータで全文取得にエスカレーション可能）とし、エージェントの文脈予算を制御する

**Step 4 — エージェントによる決定の記録**
- [ ] `kizami_record_decision` MCPツール（書き込み系）— エージェントが判断を下した直後にその場で記録できるようにする。常に `Status: Draft` として生成し、新規ファイル作成のみ（既存文書・コードの編集や削除は不可）。`kizami mcp --allow-write` で明示的にオプトインした場合のみ有効化
- [ ] `kizami hook pre-tool-use` — Claude CodeのツールフックでEdit/Writeの直前に対象ファイルを縛る `kizami context` の出力を注入する。エージェントがマニフェストを読まない・MCPツールを呼ばない場合の決定論的なフォールバックとして機能する

*複数のADRにまたがることを想定（ドッグフーディング方針を参照）。最低限、リゾルバ／注入戦略、「動詞ではなく問い」というMCPツール設計、Related Files定義の統合の3本は記録する。*

### バックログ — 候補機能

*実運用から生まれたアイデアで、まだフェーズに割り当てていないもの。*

- [ ] `kizami list --status <status>` — ステータスでリストを絞り込む（例：`--status active`）。ドキュメントが蓄積してきた際のノイズ低減に。

---

### チームフィードバック

kizami はチームで実運用中。実際の使用から得られるフィードバックがロードマップを動かす。

- GitHub Issues に `feedback` ラベルをつけて起票する
- `kizami lint` / `kizami audit` のエラーメッセージに関するフィードバックは特に価値が高い
- チーム利用中に発見した使いにくさは `docs/decisions/` に ADR として記録する（ドッグフーディング）

---

## 参考リンク

- [MADRフォーマット仕様](https://adr.github.io/madr/)
- [cobraドキュメント](https://github.com/spf13/cobra)
- [adr-tools（比較対象）](https://github.com/npryce/adr-tools)
- [GoReleaser](https://goreleaser.com/)
- [GitHub CLI（gh）](https://cli.github.com/) — ClaudeがPRを作成するために必要
