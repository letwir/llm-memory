# ISSUE-008〜010 walkthrough

## 実行計画

1. ライセンスと残Issueを確認する。
2. embedding、実機競合テスト、WORM境界を実装する。
3. ローカルおよび実DBテストを実行し、未適用migrationは未完了として記録する。

## 実施内容

- `migrations/008_embeddings.sql` とGemini embedding／pgvector CLI経路を追加した。
- `migrations/010_worm_boundary.sql` とimmutable field／DELETE拒否triggerを追加した。
- `integration` build tagの同時投入テストを追加した。

## 検証

- `go test ./...`: 成功
- `go test -tags integration -run TestConcurrentActiveIdentityInsert`: PASS（実DB）
- WORM境界テスト: `010_worm_boundary.sql` 未適用のためSKIP

## 残作業

- 008 migration適用、embedding生成、semantic検索実データ確認
- 010 migration適用後のWORM境界テスト再実行
