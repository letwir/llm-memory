# Diary: agent bootstrap prompt

### 2026-08-29 17:00:00
Task: READMEへエージェントの開始探索・終了報告プロンプトを追加する
Tried: SPR/XMLと圏論的記号を含む1行命令を追加し、skill呼び出しとExplore/Report契約を明示した
Result: README/USAGEを更新し、テストに成功した
Uncertainty: 各エージェントが追加指示を常に遵守するかは未検証
Attribution: [ワイの指示(PromptDefect): 5% {MissingSpec}] vs [AI認知(AgentDefect): 95% {Misreading}]
PromptDefect: README内の掲載位置や英日比率は指定されていなかった
AgentDefect: なし
Correction: skill案内の直後に貼り付け可能な1行と補足を配置した
Feedback: 次回は掲載位置と、指示を必須扱いにする対象エージェントを指定するとよい
Rewritten-request: READMEのエージェントskill登録節に、llm-memoryを呼び出し、開始時に探索し終了時に検証付き報告を行うSPR/XML記号付き1行プロンプトを追加し、USAGEの導線も整合させる
