# Diary: analyze recent improvements output

### 2026-08-29 16:00:00
Task: analyzeに直近の改善文章を表示する
Tried: 頻度集計とは別に新しいdiaryエントリを優先する改善文を追加し、通常表示とJSONを確認した
Result: `recent_improvements`と専用表示セクションを追加し、テストとCLI確認に成功した
Uncertainty: diaryが時系列順でない入力では「直近」の意味が弱くなる。実Hook経由は未確認。
Attribution: [ワイの指示(PromptDefect): 10% {MissingSpec}] vs [AI認知(AgentDefect): 90% {ToolFailure}]
PromptDefect: 出力形式の指定が暗黙的だった
AgentDefect: 初回実装では頻度フィードバックと直近改善文を分離していなかった
Correction: `recent_improvements`を追加し、重複排除・具体化した
Feedback: 次回は表示見出し、JSONキー、並び順、重複方針を指定するとさらに速い
Rewritten-request: analyzeの通常表示とJSONに、diaryの直近エントリ由来で重複排除した実行可能な改善文を追加し、テストとCLI出力を確認する
