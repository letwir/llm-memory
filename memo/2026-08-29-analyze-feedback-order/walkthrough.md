# Walkthrough: analyze feedback order

## 実行計画

1. Old output path and recent improvement data flowを確認する。
2. Recent actionable feedbackをprimary sectionへ移す。
3. Test and rebuild the root CLI.
4. Record and evaluate the change.

## 実施内容

- section 1を`RecentImprovements`表示へ変更。
- 旧頻度集計をsection 1.5の補足へ移動。
- USAGEの出力説明を更新。

## 検証結果

- `go test ./...`: PASS
- root `llm-mem.exe analyze`: recent feedback displayed in section 1
