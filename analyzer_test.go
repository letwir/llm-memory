// Test for analyzer.go (Causal Attribution & Feedback Generation)
package main

import (
	"strings"
	"testing"
)

func TestParseDiaryEntriesHeavy(t *testing.T) {
	sampleDiary := `
### 2026-08-14 18:00:30
Hypothesis: テスト仮説ですわ
Tried: テスト実行
Rejected: 却下案
Uncertainty: 不確実性
Attribution: [ワイの指示(PromptDefect): 60% {Underspecified}] vs [AI認知(AgentDefect): 40% {Misreading}]
Search: search.exe
Correction: 修正内容
Emotion: 最高ですわ
Thoughts: 感想ですの

### 2026-08-14 19:00:00
Hypothesis: 2件目のテスト
Tried: 実行
Attribution: [ワイの指示(PromptDefect): 20% {Contradiction}] vs [AI認知(AgentDefect): 80% {ToolFailure}]
`

	entries, err := ParseDiaryEntriesHeavy(strings.NewReader(sampleDiary))
	if err != nil {
		t.Fatalf("ParseDiaryEntriesHeavy failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// 1件目の検証
	e1 := entries[0]
	if e1.PromptRatio != 60 || e1.AgentRatio != 40 {
		t.Errorf("e1 ratio mismatch: prompt=%d, agent=%d", e1.PromptRatio, e1.AgentRatio)
	}
	if len(e1.PromptDefects) == 0 || e1.PromptDefects[0] != "Underspecified" {
		t.Errorf("e1 PromptDefects mismatch: %v", e1.PromptDefects)
	}
	if len(e1.AgentDefects) == 0 || e1.AgentDefects[0] != "Misreading" {
		t.Errorf("e1 AgentDefects mismatch: %v", e1.AgentDefects)
	}

	// 集計分析の検証
	res := AnalyzeAttributionEntriesHeavy(entries)
	if res.TotalEntries != 2 {
		t.Errorf("expected 2 total entries, got %d", res.TotalEntries)
	}
	if res.TotalPromptDefects != 2 {
		t.Errorf("expected 2 total prompt defects, got %d", res.TotalPromptDefects)
	}
	if res.TotalAgentDefects != 2 {
		t.Errorf("expected 2 total agent defects, got %d", res.TotalAgentDefects)
	}
	if res.AveragePromptRatio != 40.0 {
		t.Errorf("expected avg prompt ratio 40.0, got %f", res.AveragePromptRatio)
	}
	if len(res.UserFeedbackTips) == 0 {
		t.Errorf("expected user feedback tips, got empty")
	}
	if len(res.AgentRuleDiffs) == 0 {
		t.Errorf("expected agent rule diffs, got empty")
	}
}

func TestEdgeCasesHeavy(t *testing.T) {
	// 空テキスト
	entries, err := ParseDiaryEntriesHeavy(strings.NewReader(""))
	if err != nil {
		t.Fatalf("empty text failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty text, got %d", len(entries))
	}

	res := AnalyzeAttributionEntriesHeavy(entries)
	if len(res.UserFeedbackTips) == 0 {
		t.Errorf("expected fallback tip for empty entries")
	}

	// 不正構文テキスト
	corrupted := `
ランダムなテキスト
### 不正な日付フォーマット
hogehoge
`
	entries2, err := ParseDiaryEntriesHeavy(strings.NewReader(corrupted))
	if err != nil {
		t.Fatalf("corrupted text scan failed: %v", err)
	}
	_ = AnalyzeAttributionEntriesHeavy(entries2)
}
