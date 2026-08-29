# Diary: analyze feedback order

### 2026-08-29 18:30:00
Task: analyzeのプロンプト改善フィードバックを直近内容にする
Tried: RecentImprovementsを第一表示にし、頻度傾向を補足へ分離した
Result: root exeの実行出力で直近の改善文が第一節に表示された
Uncertainty: 配布済みの古いexeは再ビルドが必要
Attribution: [ワイの指示(PromptDefect): 5% {MissingSpec}] vs [AI認知(AgentDefect): 95% {Misreading}]
PromptDefect: primary表示と補足表示の優先順位は明示されていなかった
AgentDefect: 直前の機能追加でprimary表示を更新し忘れていた
Correction: section 1をRecentImprovements、section 1.5を頻度補足に変更した
Feedback: 表示の優先順位を要件として明記すると再発しにくい
Rewritten-request: analyzeの第一節には直近の重複排除済み改善文を表示し、過去の頻度フィードバックは補足節へ移し、テストとroot exe出力を確認する
