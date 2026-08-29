package main

import (
	"math"
	"testing"
)

func TestCalculateEvaluation(t *testing.T) {
	input := EvaluationInput{
		EvaluationID:  "eval-current",
		ComparisonKey: "verifier/code-review/v1",
		Axes: map[string]float64{
			"counterexample_quality": 0.78,
			"scope_discipline":       0.40,
		},
	}
	previous := &EvaluationPrevious{
		MemoryID: "eval-previous",
		Values: map[string]float64{
			"counterexample_quality": 0.62,
			"scope_discipline":       0.40,
		},
	}

	got := calculateEvaluation(input, previous)
	if !got.Deterministic.RangeValid || !got.Deterministic.BaselineCompatible {
		t.Fatalf("deterministic flags = %+v", got.Deterministic)
	}
	if math.Abs(got.Relative["counterexample_quality"].Delta-0.16) > 1e-9 {
		t.Fatalf("delta = %v, want 0.16", got.Relative["counterexample_quality"].Delta)
	}
	if got.Relative["counterexample_quality"].Trend != "improved" {
		t.Fatalf("trend = %q, want improved", got.Relative["counterexample_quality"].Trend)
	}
	if got.Relative["scope_discipline"].Trend != "unchanged" {
		t.Fatalf("unchanged trend = %q", got.Relative["scope_discipline"].Trend)
	}
	if got.Reuse.Eligible {
		t.Fatal("evaluation reuse must remain disabled")
	}
}

func TestCalculateEvaluationDetectsAxisDrift(t *testing.T) {
	input := EvaluationInput{
		ComparisonKey: "coder/implementation/v1",
		Axes:          map[string]float64{"correctness": 0.9},
	}
	previous := &EvaluationPrevious{
		MemoryID: "eval-previous",
		Values:   map[string]float64{"correctness": 0.8, "scope": 0.7},
	}
	got := calculateEvaluation(input, previous)
	if got.Deterministic.BaselineCompatible {
		t.Fatal("axis drift must make baseline incompatible")
	}
}

func TestEvaluationInputNormalize(t *testing.T) {
	_, err := (EvaluationInput{ComparisonKey: "key", Axes: map[string]float64{"quality": 1.1}}).normalize()
	if err == nil {
		t.Fatal("out-of-range score must fail")
	}

	got, err := (EvaluationInput{ComparisonKey: " key ", Axes: map[string]float64{"quality": 0.5}}).normalize()
	if err != nil {
		t.Fatal(err)
	}
	if got.ComparisonKey != "key" || got.EvaluationID == "" {
		t.Fatalf("normalized input = %+v", got)
	}
}

func TestTaskAttributionIsNormalizedAndIncludedInFeedback(t *testing.T) {
	input := EvaluationInput{
		TaskID:        " task-42 ",
		ComparisonKey: "task/attribution/v1",
		Axes:          map[string]float64{"completion": 1},
		Attribution: &EvaluationAttribution{
			PromptDefects: []string{" missing acceptance criteria "},
			AgentDefects:  []string{"misreading"},
		},
	}
	normalized, err := input.normalize()
	if err != nil {
		t.Fatal(err)
	}
	if normalized.TaskID != "task-42" || normalized.Attribution.PromptRatio != 0.5 || normalized.Attribution.AgentRatio != 0.5 {
		t.Fatalf("normalized attribution = %+v", normalized)
	}
	meta := calculateEvaluation(normalized, nil)
	if meta.TaskID != "task-42" || meta.Feedback == nil || meta.Feedback.Conclusion == "" {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestTaskAttributionRejectsNonUnitRatio(t *testing.T) {
	_, err := (EvaluationInput{
		TaskID:        "task-42",
		ComparisonKey: "task/attribution/v1",
		Axes:          map[string]float64{"completion": 1},
		Attribution:   &EvaluationAttribution{PromptRatio: 0.8, AgentRatio: 0.8},
	}).normalize()
	if err == nil {
		t.Fatal("non-unit attribution ratio must fail")
	}
}

func TestTaskAttributionRequiresTaskID(t *testing.T) {
	_, err := (EvaluationInput{
		ComparisonKey: "task/attribution/v1",
		Axes:          map[string]float64{"completion": 1},
		Attribution:   &EvaluationAttribution{PromptDefects: []string{"ambiguous"}},
	}).normalize()
	if err == nil {
		t.Fatal("task attribution without task_id must fail")
	}
}
