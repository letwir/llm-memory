# Diary: docs and Mermaid synchronization

### 2026-08-29 18:00:00
Task: 現行実装に合わせてdocsとGitHub Mermaid表示を修正する
Tried: DB定義・migration・README・SKILLを照合し、複合矢印を単純化した
Result: docs 3種、SKILL、README/USAGEを同期し、Goテストに成功した
Uncertainty: GitHub本番ページの描画は未確認
Attribution: [ワイの指示(PromptDefect): 5% {MissingSpec}] vs [AI認知(AgentDefect): 95% {Misreading}]
PromptDefect: どのGitHub表示環境を対象にするか未指定
AgentDefect: なし
Correction: GitHubの基本fenced Mermaidと保守的な構文に揃えた
Feedback: 次回は対象リポジトリURLと描画崩れの画面例があると再現確認まで可能
Rewritten-request: docs 3種のMermaidをGitHub互換の保守的構文へ直し、現行DB実装とREADME/SKILLを同期し、テスト後に対象変更だけコミットする
