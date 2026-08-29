# llm-memory 人間向け操作ガイド

## 初回セットアップ

1. PostgreSQLに空の `llm_memory` データベースを用意する。
2. `init.sql` を実行する。
3. `LLM_MEMORY_DB_URL` を設定する、または `build.example.ps1` で接続先を埋め込む。
4. 必要なら `BUILD_AND_INSTALL.bat` を実行して各エージェントへskillとバイナリを登録する。
5. `llm-mem status` で接続を確認する。

```bat
BUILD_AND_INSTALL.bat
```

登録先は `.gemini\skills`、`.codex\skills`、`.claude\skills`、`.agents\skills`、`.config\opencode\skills`、`.opencode\skills` です。既存の `SKILL.md` は初回更新時だけ `SKILL.md.previous` として退避します。

Goのビルドと、上記すべてのskillディレクトリへの `SKILL.md`／`llm-mem.exe` 配布を一度に行う場合は、現在の端末で `LLM_MEMORY_BUILD_DB_URL` を設定してから実行します。URLそのものは画面表示・ソース保存されませんが、ビルド済みexeには埋め込まれます。

```powershell
./BUILD_AND_INSTALL.bat
```

```powershell
$env:LLM_MEMORY_DB_URL = "<local PostgreSQL URL>"
./llm-mem.exe status
```

## ストック記憶を眺める

全カテゴリの有効な記憶を、L2→L1→L0の順で短く表示します。

```powershell
./llm-mem.exe stock -limit 20
./llm-mem.exe stock -cat knowledge -limit 50
./llm-mem.exe stock -json
```

詳細本文が必要な場合は通常検索を使います。

```powershell
./llm-mem.exe search -q "キーワード" -level 0
./llm-mem.exe search -tag "decision" -json
```

## 評価を記録する

`eval` は通常の `ingest` と別経路で、LLM抽出やグラフ化を行わず直接保存します。同じ `comparison_key` の直近評価があれば、差分と `improved` / `declined` / `unchanged` をGo側で計算します。

タスク単位の完了評価では `task_id` と `attribution` を追加します。`prompt_defects` は人間の指示の欠陥、`agent_defects` はLLMの不理解・実行失敗です。Insert成功後、CLIは結論と双方へのフィードバックを表示します。

```powershell
./llm-mem.exe eval -file .\evaluation.json
./llm-mem.exe evals -key "verifier/code-review/v1" -limit 20
```

入力例:

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
    "agent_defects": ["misreading"]
  }
}
```

現段階では評価値を保存するだけで、モデル推薦や評価値の自動再利用は行いません。

## 通常の知識取り込み

```powershell
./llm-mem.exe ingest -file .\knowledge.md -cat knowledge
./llm-mem.exe ingest -title "設計方針" -text "本文" -cat decision
./llm-mem.exe compact -limit 20
```

## diaryの因果分析

`analyze` はPromptDefect/AgentDefectの頻度集計に加えて、diaryの新しいエントリを優先した重複排除済みの `直近で改善したほうがよい指示内容` を表示します。JSON出力では `recent_improvements` として取得できます。

```powershell
./llm-mem.exe analyze -file .\diary.md
./llm-mem.exe analyze -file .\diary.md -json
```

## 変更・公開時の注意

- DB接続URLを含むexe、`.env`、評価実データはGitへ追加しない。
- `init.sql` は既存DBへ適用する前にバックアップする。
- `eval` の評価イベントは追記型で、既存評価を上書きしない。
- `GEMINI_*_API_KEY` はLLM抽出を使う場合だけ必要で、リポジトリへ保存しない。
