# ワイの使ってるLLM外部脳です。

`llm-memory` は、LLMとの作業記憶・知識・評価履歴をPostgreSQLへ保存する、個人運用向けのCLIです。
通常の知識取り込み、評価イベントの相対比較、知識グラフをひとつの外部脳として扱えます。

# llm-memory (LLM Multi-Client Bi-Temporal Memory CLI)

[English](README.md) | [日本語](README.md)

## ナニコレ？
Windows 4台やDebianからTailscale経由で同時に読み書き・自己編集できる、PostgreSQL 18バックエンドのエージェント記憶基盤CLIです。  
JITMINDの自己編集ライフサイクル（ADD/UPDATE/NOOP）とGraphitiの多段縮約（L0〜L3）・ナレッジグラフ抽出を統合し、愚痴・因果分析から設計方針までを自動で構造化蓄積します。

## 必要なもの
- **PostgreSQL 18+**: `llm_memory` データベースおよび `llm` ユーザー（リモートDebianサーバー等）
- **Go 1.22+**: ビルド用（単一静的バイナリ `llm-mem.exe`）
- **環境変数**:
  - `LLM_MEMORY_DB_URL`: PostgreSQL接続URL（実値は環境変数など安全な場所で設定）
  - `GEMINI_GROUNDING_API_KEY` または `GEMINI_API_KEY`（任意・推奨: 高精度LLM抽出に利用）

## DB接続先についての重要な注意

DB接続先は通常 `LLM_MEMORY_DB_URL` 環境変数で渡せます。さらに、配布先で環境変数を設定せずに使う場合は、ビルド時に `-ldflags -X main.buildDatabaseURL=...` でexe内部へ埋め込めます。

これは便利な反面、接続先だけでなくURLに含めた認証情報もexeから回収可能になります。埋め込み済みexeをGitHubへコミットしたり、第三者へ配布したりしないでください。公開リポジトリには接続先の実値・認証情報を置かず、`build.example.ps1` を参考に手元でビルドしてください。

```powershell
$env:LLM_MEMORY_BUILD_DB_URL = "<local PostgreSQL URL>"
./build.example.ps1 -Output llm-mem.exe
```

詳細な人間向け操作手順は [USAGE.md](USAGE.md) を参照してください。

## エージェントskillとして登録

Windowsでは `INSTALL.bat` を実行すると、Gemini / Codex / Claude / 汎用 `.agents` / OpenCode向けのskillディレクトリへ `SKILL.md` を登録します。

```bat
INSTALL.bat
```

`LLM_MEMORY_BIN` と `LLM_MEMORY_HOME` もユーザー環境変数へ設定されるため、実行後は新しいターミナルを開いてください。既存のskillは削除せず、初回更新時に `.previous` として退避します。

## SQL初期化

空のPostgreSQLデータベースを作成した後、`init.sql` を一度実行します。`pgcrypto` と `pg_trgm` の拡張、テーブル、アクティブビュー、検索用インデックスを作成します。

```powershell
psql "$env:LLM_MEMORY_ADMIN_URL" -c "CREATE DATABASE llm_memory"
psql "$env:LLM_MEMORY_DB_URL" -f .\init.sql
```

既存データベースへ実行する場合は、事前にバックアップを取得してください。`CREATE TABLE IF NOT EXISTS` と `CREATE INDEX IF NOT EXISTS` は既存テーブルの列定義を修正しません。

## 使い方

### 1. 統合ナレッジ取り込み (`ingest`) ★推奨
ファイル（`diary.md`, `decisions.md`, `Walkthrough.md` 等）やテキストを投入すると、**自己編集判定 ＋ 多段縮約 ＋ 知識グラフ抽出** を一括実行します。

```powershell
$bin = ".\llm-mem.exe"

# ファイルから自動カテゴリ推定取り込み
& $bin ingest -file "path/to/diary.md"
& $bin ingest -file "path/to/decisions.md"
& $bin ingest -file "path/to/Walkthrough.md"

# 直接テキストを取り込み
& $bin ingest -title "新設計方針" -text "本文..." -cat "knowledge"
```

### 評価イベント (`eval` / `evals`)

評価は通常の `ingest` と分離し、LLM抽出・グラフ化・supersedeを経由せずに保存します。
`eval` は同じ `comparison_key` の直近評価との差分をGo側で計算して新規Insertし、`evals` は履歴参照だけを行います。評価の再利用・モデル推薦は現段階では無効です。

```powershell
& $bin eval -file "evaluation.json"
& $bin evals -key "verifier/code-review/v1" -limit 20 -json
```

入力JSONの最小形式:

```json
{
  "task_id": "task-2026-08-29-example",
  "comparison_key": "verifier/code-review/v1",
  "role": "verifier",
  "persona": "adversarial",
  "axes": {
    "counterexample_quality": 0.78,
    "scope_discipline": 0.92
  },
  "attribution": {
    "prompt_defects": ["missing acceptance criteria"],
    "agent_defects": ["misreading"],
    "prompt_ratio": 0.5,
    "agent_ratio": 0.5
  }
}
```

タスク終了時の評価は、最後の知識Insert後に実行します。結果は同じ `memories` テーブルへ `category=eval` として追記され、`metadata.feedback` に人間の指示へのフィードバックとLLM側の修正点が保存されます。通常表示では `タスク評価リザルト` としてその場で出力されます。

タスク記録は `./memo/<task_id>/` に `knowledge.md`、`diary.md`、`walkthrough.md` の3ファイルを作成します。`walkthrough.md` には `実行計画`、実施内容、変更ファイル、検証結果、残作業を含めます。

### 2. 記憶の横断検索 (`search`)
```powershell
# キーワード検索 (L1要点表示)
& $bin search -q "JITMIND" -level 1

# カテゴリ別検索 (diaryの愚痴・反省・因果比率を1行表示)
& $bin search -cat "diary" -level 2

# タグ検索 & JSON構造化出力
& $bin search -tag "postgres" -json

# Memory Object の種別・適用範囲で絞り込み
& $bin search -type decision -scope project -json

# 正規化された長期記憶を直接登録
& $bin add -title "JSONBを採用" -content "分析データはPostgreSQL JSONBへ保存する" `
  -type decision -scope project -rationale "DDL churnを避ける" `
  -evidence "decisions.md#DECISION-001" -rejected "FLAC tagsへ全配列, featureごとのDDL" -confidence 0.95
```

`add`/`ingest`/`supersede` は `metadata.memory_object` に、種別・scope・命題・根拠・理由・却下案・確信度を正規化して保存します。既存の `category` や L0〜L3 は互換維持され、Memory Object の type を省略した場合はカテゴリから補完されます。

### 3. 未縮約レコードのバッチ多段要約 (`compact`)
```powershell
& $bin compact -limit 20
```

### 4. 知識グラフの確認 (`graph`)
```powershell
& $bin graph list
```

### 5. 疎通確認 & 端末一覧 (`status` / `clients`)
```powershell
& $bin status
& $bin clients
```

## 概要詳しく

各状態ファイルに応じた詳細な機能・仕様は以下のドキュメントをご参照ください。

| ドキュメント | 概要 | リンク |
|---|---|---|
| **状態遷移図** | JITMIND自己編集ライフサイクル（ADD / UPDATE / NOOP / DEPRECATED） | [state_diagram.md](docs/state_diagram.md) |
| **ER図とデータ構造** | PostgreSQL 18 のテーブル定義（clients, memories, nodes, edges） | [database_er_diagram.md](docs/database_er_diagram.md) |
| **アーキテクチャ & 縮約** | Multi-client Tailscale 接続と L0〜L3 多段縮約パイプライン | [shm_architecture.md](docs/shm_architecture.md) |

## 状態図
二重時間軸（Valid Time / Transaction Time）における記憶の自己編集・バージョン昇格のライフサイクルを定義しています。  
詳細は [docs/state_diagram.md](docs/state_diagram.md) をご覧ください。

## ER図とデータ構造
端末管理 (`clients`)、二重時間軸記憶 (`memories`)、および知識グラフ (`knowledge_nodes`, `knowledge_edges`) のリレーショナル構造です。  
詳細は [docs/database_er_diagram.md](docs/database_er_diagram.md) をご覧ください。

## SHM/WORM
多端末からの同時参照に耐えうるTrigramインデックス全文検索と、追記型・二重時間軸による改ざん不能な不変履歴（WORM構造）を両立しています。  
詳細は [docs/shm_architecture.md](docs/shm_architecture.md) をご覧ください。

## ライセンス
MIT License / Proprietary to LTW Ecosystem.  
**Warning**: 本ツールの本番環境接続情報（Tailscale IP・DB認証情報）は `.env` や環境変数で管理し、Gitリポジトリへコミットしないでください。
