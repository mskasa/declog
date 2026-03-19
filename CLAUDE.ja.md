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
│       ├── 0001-use-go-over-shell-script.md
│       ├── 0002-use-cobra-for-cli-framework.md
│       ├── 0003-madr-format-compatibility.md
│       ├── 0004-plaintext-markdown-only.md
│       └── 0005-ripgrep-fallback-strategy.md
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
kizami list                       # 新しい順に一覧表示（ID・日付・ステータス・タイトル）
kizami search <keyword>           # キーワード検索
kizami show <id>                  # 指定IDのDecisionを表示（例: kizami show 3）
kizami status <id> <status>       # ステータス変更（例: kizami status 3 superseded --by 5）
kizami blame <file>               # 指定ファイルに関連するDecisionを逆引き
```

### ステータス定義

| ステータス           | 意味                               | 使うタイミング                             |
| -------------------- | ---------------------------------- | ------------------------------------------ |
| `Active`             | 現在有効な判断（デフォルト）       | コード変更と同時にコミット                 |
| `Inactive`           | 単純に無効になった                 | 置き換え先のADRが存在しない場合            |
| `Superseded by NNNN` | 別のADRに置き換えられた            | 新しいADRを作成した場合                    |

**ステータス運用方針：**
- デフォルトは `Active`。ADRはコード変更と同時にコミットする運用のため、作成時点で意思決定済みとみなす
- 設計を覆す新しいADRを作成した場合は既存ADRを `Superseded by NNNN` にする
- 置き換え先のADRが存在しない場合は `Inactive` にする

---

## Markdownテンプレート（MADR準拠）

`kizami adr` 実行時に生成されるテンプレート：

```markdown
# {NNNN}: {Title}

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
NNNN-kebab-case-title.md
例: 0001-use-go-over-shell-script.md
```

- `NNNN`：4桁ゼロ埋め連番（既存ファイルの最大値＋1で自動採番）
- kebab-case：タイトルを小文字・ハイフン区切りに自動変換
- 保存先：`docs/decisions/`（リポジトリルートからの相対パス）
- このリポジトリのドッグフーディング用ADRは、英語版と日本語版の両方を作成する：
  - 英語版：`docs/decisions/0001-use-go-over-shell-script.md`
  - 日本語版：`docs/decisions/ja/0001-use-go-over-shell-script.md`

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
// Decision to use git config instead of an environment variable: docs/decisions/0009-author-source.md
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
- StatusをActive → Inactive または Superseded by NNNN に変更する
- 誤字脱字の修正
- Related Filesへの追記

**Supersededを使うケース：**
- 判断の方向性ごと変わった場合は新しいADRを作成してSupersededにする
- 同じ判断の修正・更新であれば直接更新で構わない

**ADR更新時のコミットメッセージ：**
- 何をなぜ変えたかを明記する
- 例：`docs: update ADR 0003 - increase pool size from 10 to 20 based on load test`
- 悪い例：`update adr`

### 開発開始時点で作成すべき初期ADR

コードを1行も書く前に、以下のADRを手動で作成しておくこと：

| ID   | タイトル                    | 内容                                                          |
| ---- | --------------------------- | ------------------------------------------------------------- |
| 0001 | use-go-over-shell-script    | Goを選んだ理由（シングルバイナリ、Windows対応、型安全）       |
| 0002 | use-cobra-for-cli-framework | cobraを選んだ理由（デファクト、シェル補完、サブコマンド管理） |
| 0003 | madr-format-compatibility   | MADRフォーマット準拠の理由（既存ADRツールとの互換性）         |
| 0004 | plaintext-markdown-only     | DBを使わずMarkdownのみにした理由（Git親和性、可搬性）         |
| 0005 | ripgrep-fallback-strategy   | ripgrep依存とフォールバック設計の判断                         |
| 0006 | command-name-why            | コマンド名をもともと `why` にした理由（0010 で superseded）   |

---

## 開発ルール

### コーディング規約

- コミット前に必ず `gofmt` / `goimports` を通す
- エラーはwrapする（`fmt.Errorf("...: %w", err)`）
- CLIの出力メッセージは**英語**に統一する
- コードコメントは**英語**に統一する

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
「docs/decisions/0007-auto-numbering-strategy.md を作成しました。
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

## 実装状況

<!-- 作業が進むたびにここを更新する -->

### MVP (v0.1.0) ✅

- [x] .github/workflows/ci.yml（PR毎にgo test + go vet）
- [x] go.mod + cobraセットアップ（`module github.com/mskasa/kizami`）
- [x] cmd/root.go（`kizami` コマンドのルート）
- [x] internal/decision/generate.go（自動採番・ファイル生成）
- [x] internal/template/template.go（Markdownテンプレート）
- [x] cmd/log.go（`kizami adr` / `kizami design`）
- [x] cmd/list.go（`kizami list`）
- [x] cmd/search.go（`kizami search`）
- [x] cmd/show.go（`kizami show`）
- [x] cmd/status.go（`kizami status`）
- [x] docs/decisions/ 初期ADR（0001〜0006）
- [x] README.md
- [x] GoReleaser設定

### v0.1.0（残り）

- [x] ロゴ画像作成
- [x] cmd/blame.go（`kizami blame <file>` — ADR内のファイルパス記述を全文検索）
- [x] `kizami --version`

### v0.2.0

- [x] `kizami init`
- [x] `kizami adr` のエディタ自動起動
- [x] `kizami adr` 実行時にステージング済み・未ステージングの両方の変更ファイルを候補としてRelated Filesに提示する
- [x] `kizami adr` 実行時の類似ADR提示（キーワード部分一致）
- [x] `kizami list --status`
- [x] `kizami supersede`
- [x] `kizami review`（長期未更新ADRの検出）
- [x] git hookでADR追加を促す仕組み
- [x] GitHub Actions連携（`kizami init` でワークフロー生成）

### v0.3.0

- [x] `kizami audit`（Related Filesのコードとの乖離検出）
- [x] `kizami audit` のCI定期実行（週次・GitHub Issue自動作成）
- [x] LLM連携によるADRドラフト自動生成
- [x] `kizami init` 実行時に `~/.config/kizami/config.toml` をデフォルト値で生成する

### kizami へのリネーム ✅

- [x] GitHub リポジトリのリネーム: `mskasa/declog` → `mskasa/kizami`
- [x] `go.mod` モジュールパスの更新: `github.com/mskasa/declog` → `github.com/mskasa/kizami`
- [x] コードベース全体のインポートパス更新
- [x] バイナリ名の変更: `why` → `kizami`（cmd/root.go, .goreleaser.yaml）
- [x] 設定パスの変更: `~/.config/declog/` → `~/.config/kizami/`
- [x] README.md / README.ja.md の更新
- [x] CLAUDE.md / CLAUDE.ja.md の更新（新アイデンティティを反映）
- [x] `why` コマンドを参照している既存 ADR の更新

### v0.4.0（スコープ拡張）

- [x] `kizami adr` / `kizami design` — 作成コマンドの分離（`kizami log --type` の代替）
- [x] 設計書テンプレートの追加（保存先 `docs/design/`、デフォルト `Status: Draft`）
- [x] ADR テンプレートのデフォルトを `Status: Active` → `Status: Draft` に変更
- [x] `kizami audit` で `Draft` ドキュメントをスキップ（`Active` のみ対象）
- [x] `kizami init` にオプションで auto-promote ワークフロー（`kizami-promote.yml`）を生成する機能を追加：main へのプッシュ時に `Draft` → `Active` へ自動昇格、カスタマイズ用コメント付き
- [x] `kizami audit` で複数ディレクトリをスキャン可能に（config の `audit.dirs`）
- [x] 汎用メッセージから ADR 固有の表現を除去
- [ ] `kizami design --ai` — 設計書向け AI ドラフト生成
- [x] golangci-lint を CI に追加
- [x] mise ツールチェーン設定（Go と golangci-lint のバージョンをローカル環境でも固定）
- [ ] cmd/ パッケージのテスト追加
- [ ] Related Files にディレクトリパスを指定可能にする（配下のファイルすべてが関連ファイルとして扱われる）
- [ ] このリポジトリ自体に `kizami init` を実行する（ドッグフーディング）
- [ ] このリポジトリ自体に設計書を作成する（ドッグフーディング）

### v1.0.0（パブリックリリース）

- [ ] ドキュメントサイト（GitHub Pages）
- [ ] Homebrew formula
- [ ] カラー出力（kizami list / kizami search）

### バックログ

- [ ] ファイル存在確認を超えた乖離検出（関数名・シンボルレベルの参照チェック）
- [ ] 逆引きインデックスの生成（`.kizami/index.json`：ファイルパス → ADR ID のマッピング）による `kizami blame` の高速化と外部ツール連携
- [ ] `kizami sync` — 既存ドキュメントの Related Files を対話的に更新
- [ ] テンプレートのユーザー定義（config でテンプレートパスを指定可能に。Related Files セクションの必須化については要検討）
- [ ] kizami stats
- [ ] GitHub Actions Marketplace 公開

---

## 参考リンク

- [MADRフォーマット仕様](https://adr.github.io/madr/)
- [cobraドキュメント](https://github.com/spf13/cobra)
- [adr-tools（比較対象）](https://github.com/npryce/adr-tools)
- [GoReleaser](https://goreleaser.com/)
- [GitHub CLI（gh）](https://cli.github.com/) — ClaudeがPRを作成するために必要
