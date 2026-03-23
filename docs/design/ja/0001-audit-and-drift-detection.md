# 0001: Audit と Drift Detection

- Date: 2026-03-23
- Type: Design
- Status: Active
- Author: masahiro.kasatani

## Overview

`kizami audit` は、ドキュメントの `## Related Files` セクションに記載されたソースファイルがリポジトリ上に存在しなくなったことを検出します。
これにより、開発者が手動で更新することを覚えていなくても、陳腐化した参照を自動的に発見し、ドキュメントの正確性を維持します。

## Background

kizami のコアバリューは、設計判断・設計書とそれが説明するコードの間にトレーサビリティを維持することです。
そのリンクは、すべてのドキュメントの `## Related Files` セクションで表現されます。

自動チェックがなければ、このリンクは静かに劣化していきます。

- ファイルがリネームされた → ドキュメントは古いパスを指したまま
- モジュールが削除された → ドキュメントは存在しないコードを説明したまま
- ディレクトリが再編された → ドキュメントの参照が誤ったものになる

開発者がファイルを移動する際にドキュメントを更新することはほとんどありません。
`kizami audit` はこのドリフトを、恒久的な問題になる前に可視化します。

## Goals / Non-Goals

**Goals:**
- `## Related Files` に記載されたファイルパス・ディレクトリパスで、存在しなくなったものを検出する
- CLI コマンド（`kizami audit`）としても、スケジュール CI ジョブとしても実行可能にする
- 複数のドキュメントディレクトリをサポートする（ADR + 設計書）
- Draft ドキュメントはスキップする（作成途中であり、まだ正式なものではない）

**Non-Goals:**
- シンボルレベルのドリフト検出（例：ファイル内で関数名がリネームされた）— ファイルの存在確認が検出の境界
- 陳腐化した参照の自動更新 — audit は報告するだけ。修正は人間が判断する
- `audit.dirs` で設定されたディレクトリ外のドキュメントのチェック

## Design

### Related Files メカニズム

kizami が作成するすべてのドキュメントには `## Related Files` セクションが含まれます。

```markdown
## Related Files

- `internal/decision/audit.go`
- `internal/search/blame.go`
- `cmd/audit.go`
```

このセクションがドキュメントとコードの結びつきの正式な記録です。
`ParseRelatedFiles`（`internal/decision/audit.go`）は行単位でパースします。

1. `## Related Files` 見出しが見つかるまでスキャン
2. リスト項目（`- path` または `- \`path\``）を収集
3. 次の `##` 見出しで停止

### ディレクトリプレフィックスエントリ

末尾が `/` のパスはディレクトリエントリとして扱われ、そのパス配下のすべてのファイルにマッチします。

```markdown
## Related Files

- `internal/search/`
```

これは「このドキュメントは `internal/search/` 配下のすべてに関連する」という意味です。

- **`kizami audit` の場合**: ディレクトリパス自体を `os.Stat` でチェックします。`internal/search/` が丸ごと削除されれば、audit が missing として報告します。
- **`kizami blame <file>` の場合**: blame はさらにディレクトリエントリのマッチも行います。`internal/search/search.go` を検索すると、`internal/search/` をリストしているドキュメントも発見されます（`internal/search/blame.go` の `blameDirEntries` 参照）。

この規約はスケーラブルです。サブシステム全体を扱うドキュメントは、ファイルを個別に列挙する代わりにディレクトリ全体を参照できます。

### ドリフト検出アルゴリズム

`internal/decision/audit.go` の `Audit(dir, repoRoot string)`:

```
1. dir 内のすべてのドキュメントを一覧取得（ID 順）
2. 各ドキュメントについて:
   a. Status != "Active"（大文字小文字無視）なら스킵
   b. Related Files エントリをパース
   c. エントリがなければスキップ
   d. 各エントリについて: os.Stat(filepath.Join(repoRoot, entry))
   e. os.IsNotExist(err) == true のエントリを収集
3. missing が1件以上あるドキュメントを AuditResult{Decision, MissingFiles} として返す
```

**なぜ Active ドキュメントのみ対象にするか？**
Draft ドキュメントは作成中です。Related Files はまだ作成されていないファイルへの予定パスや、探索的な参照を含む場合があります。これらを audit するとノイズが多くなります。Active（正式な）ドキュメントのみが正確性の基準に従います。

### マルチディレクトリサポート

`kizami audit` は `[audit] dirs`（デフォルト: `docs/decisions/` と `docs/design/` の両方）で設定されたすべてのディレクトリを走査します。

```
cmd/audit.go:
  dirs := auditDirs(root, cfg)
  for _, dir := range dirs:
    results += Audit(dir, root)
```

`audit.dirs` の設定は `documents.dirs` にフォールバックするため、明示的な設定なしに ADR と設計書の両方がカバーされます。

### CI インテグレーション

`kizami init` が生成する `.github/workflows/adr-audit.yml` は以下を行います。

1. 週次スケジュール（`cron: '0 0 * * 1'`）で `kizami audit` を実行
2. `workflow_dispatch` による手動トリガーもサポート
3. 陳腐化した参照が見つかった場合（`stale_found` output 経由）、完全な audit レポートを含む GitHub Issue を作成
4. Issue の重複防止: `[kizami audit]` の Issue は常に1件のみオープン

### Blame: 逆引き検索

`kizami blame <file>` は補完的な問いに答えます。「このファイルに言及しているドキュメントはどれか？」

ドキュメントディレクトリに対して2パスで処理します。

1. **全文検索**（ripgrep または stdlib フォールバック）: 正確なファイルパス文字列を含むドキュメントを検索
2. **ディレクトリプレフィックスマッチ**（`blameDirEntries`）: 問い合わせファイルパスのプレフィックスとなるディレクトリエントリを持つドキュメントを検索

結果はファイルパスで重複排除し、ドキュメント ID 順にソートされます。
これは audit の逆操作です。audit は Related Files が消えたドキュメントを探し、blame はまだ存在するファイルに対応するドキュメントを探します。

## Open Questions

- **シンボルレベルのドリフト**: ドキュメントで参照されている関数がリネームされてもファイルが残っている場合、audit は検出できません。将来的なアプローチとして、ドキュメント本文から関数名をパースし、AST や `ctags` 出力と照合することが考えられます。
- **ファイルのリネーム**: `git mv` はファイルをリネームしますが、kizami は git 履歴を認識しません。将来的な `kizami sync` コマンドが `git log --follow` を利用してリネームを検出し、Related Files の更新を提案できる可能性があります。

## Related Files

- `internal/decision/audit.go`
- `internal/search/blame.go`
- `cmd/audit.go`
- `internal/initializer/templates/adr-audit.yml`
- `kizami.toml`
