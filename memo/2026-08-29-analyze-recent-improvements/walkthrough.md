# Walkthrough: analyze recent improvements output

## 実行計画

1. 現行analyzeの出力責務とテストを確認する。
2. 直近改善文を通常表示・JSONへ追加する。
3. テストとCLI出力を確認する。
4. 記録と評価をllm-memへ登録する。

## 実施内容

- `AttributionAnalysisResult`へ`recent_improvements`を追加。
- diaryを新しい順に処理し、重複を除去した改善文を生成。
- 通常表示へ「直近で改善したほうがよい指示内容」セクションを追加。
- `USAGE.md`へ利用方法を追記。

## 検証結果

- `go test ./...`: PASS
- CLI出力: 対象範囲・入力形式・出力形式・完了条件、前提コンテキストを表示

## 残作業

- 実Hook経由の表示確認は未実施。
