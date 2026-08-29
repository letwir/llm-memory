# Walkthrough: agent bootstrap prompt

## 実行計画

1. READMEのskill登録案内とllm-memoryの呼び出し契約を確認する。
2. 探索と最終報告を強制する1行プロンプトを追加する。
3. USAGEの削除済みINSTALL参照を更新する。
4. テストと記録登録を確認する。

## 実施内容

- READMEにSPR/XML・圏論的記号付きの1行指示を追加。
- `BUILD_AND_INSTALL.bat`を正規導線としてREADME/USAGEを整合。

## 検証結果

- `go test ./...`: PASS（リポジトリ内GOCACHE使用）
- `git diff --check`: PASS
- 実エージェントでのプロンプト遵守: 未検証
