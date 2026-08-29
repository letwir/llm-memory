# Walkthrough: docs and Mermaid synchronization

## 実行計画

1. 現行実装とdocsの不一致を確認する。
2. MermaidをGitHub向けに簡素化する。
3. README/USAGE/SKILLとDB図・状態図・構成図を同期する。
4. テスト、差分、コミット範囲を確認する。

## 実施内容

- `SKILL.md`の削除済みINSTALL参照をBUILD_AND_INSTALLへ更新。
- ER図へuuid、search_document、memory_embeddingsを追加。
- 構成図へ全文検索、pgvector HNSW、WORM triggerを追加。
- 状態図の`valid_to=NULL`と非推奨化の説明を修正。

## 検証結果

- `go test ./...`: PASS
- `git diff --check`: PASS
- GitHub公式Mermaid埋め込み方式を確認
- 変更対象のみコミット予定

## 残作業

- GitHub上の実描画確認は未実施。
