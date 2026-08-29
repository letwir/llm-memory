# Diary: build and install batch

### 2026-08-29 16:30:00
Task: Goのビルドと各SKILL位置への配布を一発化する
Tried: 既存INSTALL.batの削除状態を尊重し、自己完結batchを追加した
Result: `BUILD_AND_INSTALL.bat`を追加し、pwsh build・6配置先へのSKILL/exe配布・失敗時停止を実装した
Uncertainty: 実配布は未実行。各exeのロックと書込み権限は利用者環境で要確認。
Attribution: [ワイの指示(PromptDefect): 10% {MissingSpec}] vs [AI認知(AgentDefect): 90% {Misreading}]
PromptDefect: 配布対象と実行条件の詳細は暗黙的だった
AgentDefect: 最初に削除済みINSTALL.batを復元しかけた
Correction: INSTALL.bat非依存の自己完結batchへ修正した
Feedback: 配布対象、必要環境、dry-run要否を指定するとさらに安全
Rewritten-request: INSTALL.batは使わず、LLM_MEMORY_BUILD_DB_URLを要求し、pwshでGoをbuild、SKILL.mdとllm-mem.exeを6つのskill pathへ配布するbatchを追加し、guardとunit testを確認する
