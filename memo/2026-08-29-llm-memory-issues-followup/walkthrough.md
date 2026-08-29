# llm-memory Issues follow-up walkthrough

## 実行計画

1. Issues.mdと実装状態を確認する。
2. Issues.mdおよびtask記録をllm-memへ取り込む。
3. 残Issueがある場合はIssues.mdを保持する。
4. テストと差分を確認し、変更をcommitする。

## 実施内容

- transaction-time、部分失敗可視化、検索拡張、競合境界、文書・ライセンス整合を実装した。
- PostgreSQL実機でsearch_document trigger、661件バックフィル、GIN indexを適用した。
- rule/walkthroughの完全一致重複を統合し、knowledgeの別内容2件はタイトルを分離した。
- Issues.mdにはembedding、実機競合テスト、WORM境界テストが残っているため削除しない。

## 検証

- `go test ./...`: 成功
- `git diff --check`: 成功
- DB側のタイトル確認、重複確認、一意index作成: ユーザー実行で成功

## 残作業

- embedding/pgvectorの導入設計
- PostgreSQL実機での競合投入テスト
- WORM境界の実機検証
