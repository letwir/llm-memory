package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const evaluationRecordType = "eval"

// EvaluationInput は、今回の評価として外部から受け取る値です。
// 数値の採点は呼び出し側が行いますが、前回との差分はGo側で計算します。
type EvaluationInput struct {
	EvaluationID  string                 `json:"evaluation_id,omitempty"`
	TaskID        string                 `json:"task_id,omitempty"`
	ComparisonKey string                 `json:"comparison_key"`
	Role          string                 `json:"role,omitempty"`
	Persona       string                 `json:"persona,omitempty"`
	TaskFamily    string                 `json:"task_family,omitempty"`
	ContractHash  string                 `json:"contract_hash,omitempty"`
	Axes          map[string]float64     `json:"axes"`
	EvidenceRefs  []string               `json:"evidence_refs,omitempty"`
	Attribution   *EvaluationAttribution `json:"attribution,omitempty"`
}

// EvaluationAttribution records whether a task failure was caused by the
// human instruction (PromptDefect) or by the LLM (AgentDefect).
type EvaluationAttribution struct {
	PromptDefects []string `json:"prompt_defects,omitempty"`
	AgentDefects  []string `json:"agent_defects,omitempty"`
	PromptRatio   float64  `json:"prompt_ratio,omitempty"`
	AgentRatio    float64  `json:"agent_ratio,omitempty"`
}

type EvaluationFeedback struct {
	Conclusion      string   `json:"conclusion"`
	UserFeedback    []string `json:"user_feedback,omitempty"`
	AgentCorrection []string `json:"agent_correction,omitempty"`
}

type EvaluationRelative struct {
	Delta float64 `json:"delta"`
	Trend string  `json:"trend"`
}

type EvaluationPrevious struct {
	MemoryID string             `json:"memory_id"`
	Values   map[string]float64 `json:"values"`
}

type EvaluationMetadata struct {
	RecordType    string                        `json:"record_type"`
	SchemaVersion string                        `json:"schema_version"`
	EvaluationID  string                        `json:"evaluation_id"`
	TaskID        string                        `json:"task_id,omitempty"`
	ComparisonKey string                        `json:"comparison_key"`
	Role          string                        `json:"role,omitempty"`
	Persona       string                        `json:"persona,omitempty"`
	TaskFamily    string                        `json:"task_family,omitempty"`
	ContractHash  string                        `json:"contract_hash,omitempty"`
	Axes          map[string]float64            `json:"axes"`
	Previous      *EvaluationPrevious           `json:"previous,omitempty"`
	Relative      map[string]EvaluationRelative `json:"relative,omitempty"`
	Deterministic EvaluationDeterministic       `json:"deterministic"`
	EvidenceRefs  []string                      `json:"evidence_refs,omitempty"`
	Attribution   *EvaluationAttribution        `json:"attribution,omitempty"`
	Feedback      *EvaluationFeedback           `json:"feedback,omitempty"`
	Reuse         EvaluationReuse               `json:"reuse"`
}

type EvaluationDeterministic struct {
	RangeValid         bool `json:"range_valid"`
	BaselineCompatible bool `json:"baseline_compatible"`
}

type EvaluationReuse struct {
	Eligible bool `json:"eligible"`
}

func (in EvaluationInput) normalize() (EvaluationInput, error) {
	in.TaskID = strings.TrimSpace(in.TaskID)
	in.ComparisonKey = strings.TrimSpace(in.ComparisonKey)
	if in.ComparisonKey == "" {
		return in, fmt.Errorf("comparison_keyは必須でしてよ")
	}
	if len(in.Axes) == 0 {
		return in, fmt.Errorf("axesは1件以上必要でしてよ")
	}
	if in.EvaluationID == "" {
		in.EvaluationID = fmt.Sprintf("eval-%d", time.Now().UTC().UnixNano())
	}
	for name, value := range in.Axes {
		if strings.TrimSpace(name) == "" {
			return in, fmt.Errorf("評価軸名が空ですわ")
		}
		if value < 0 || value > 1 {
			return in, fmt.Errorf("評価軸 %q の値 %.6f は0以上1以下で指定してくださいまし", name, value)
		}
	}
	if in.Attribution != nil {
		if in.TaskID == "" {
			return in, fmt.Errorf("attribution付きのタスク評価にはtask_idが必須でしてよ")
		}
		in.Attribution.PromptDefects = trimEvaluationItems(in.Attribution.PromptDefects)
		in.Attribution.AgentDefects = trimEvaluationItems(in.Attribution.AgentDefects)
		if in.Attribution.PromptRatio < 0 || in.Attribution.PromptRatio > 1 || in.Attribution.AgentRatio < 0 || in.Attribution.AgentRatio > 1 {
			return in, fmt.Errorf("PromptDefect/AgentDefect比率は0以上1以下で指定してくださいまし")
		}
		if in.Attribution.PromptRatio == 0 && in.Attribution.AgentRatio == 0 {
			total := len(in.Attribution.PromptDefects) + len(in.Attribution.AgentDefects)
			if total > 0 {
				in.Attribution.PromptRatio = float64(len(in.Attribution.PromptDefects)) / float64(total)
				in.Attribution.AgentRatio = float64(len(in.Attribution.AgentDefects)) / float64(total)
			}
		}
		if sum := in.Attribution.PromptRatio + in.Attribution.AgentRatio; sum > 0.000001 && (sum < 0.999999 || sum > 1.000001) {
			return in, fmt.Errorf("PromptDefect/AgentDefect比率の合計は1で指定してくださいまし")
		}
	}
	return in, nil
}

func trimEvaluationItems(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func previousEvaluation(rec *MemoryRecord) (*EvaluationPrevious, error) {
	if rec == nil {
		return nil, nil
	}
	values := make(map[string]float64)
	rawAxes, ok := rec.Metadata["axes"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("前回評価 %s のaxes形式が不正でしてよ", rec.ID)
	}
	for name, raw := range rawAxes {
		value, ok := raw.(float64)
		if !ok {
			return nil, fmt.Errorf("前回評価 %s の評価軸 %q が数値ではありませんわ", rec.ID, name)
		}
		values[name] = value
	}
	return &EvaluationPrevious{MemoryID: rec.ID, Values: values}, nil
}

func calculateEvaluation(in EvaluationInput, previous *EvaluationPrevious) EvaluationMetadata {
	meta := EvaluationMetadata{
		RecordType:    evaluationRecordType,
		SchemaVersion: "eval.v1",
		EvaluationID:  in.EvaluationID,
		TaskID:        in.TaskID,
		ComparisonKey: in.ComparisonKey,
		Role:          in.Role,
		Persona:       in.Persona,
		TaskFamily:    in.TaskFamily,
		ContractHash:  in.ContractHash,
		Axes:          in.Axes,
		Previous:      previous,
		Relative:      make(map[string]EvaluationRelative),
		Deterministic: EvaluationDeterministic{
			RangeValid:         true,
			BaselineCompatible: previous != nil,
		},
		EvidenceRefs: in.EvidenceRefs,
		Reuse:        EvaluationReuse{Eligible: false},
	}
	if in.Attribution != nil {
		meta.Attribution = in.Attribution
		meta.Feedback = buildEvaluationFeedback(in.Attribution)
	}
	if previous == nil {
		return meta
	}
	for name, current := range in.Axes {
		old, exists := previous.Values[name]
		if !exists {
			meta.Deterministic.BaselineCompatible = false
			continue
		}
		delta := current - old
		trend := "unchanged"
		if delta > 0 {
			trend = "improved"
		} else if delta < 0 {
			trend = "declined"
		}
		meta.Relative[name] = EvaluationRelative{Delta: delta, Trend: trend}
	}
	for name := range previous.Values {
		if _, exists := in.Axes[name]; !exists {
			meta.Deterministic.BaselineCompatible = false
		}
	}
	return meta
}

func buildEvaluationFeedback(attr *EvaluationAttribution) *EvaluationFeedback {
	if attr == nil {
		return nil
	}
	feedback := &EvaluationFeedback{}
	switch {
	case attr.PromptRatio > attr.AgentRatio:
		feedback.Conclusion = "主因は人間の指示側（PromptDefect）です。"
	case attr.AgentRatio > attr.PromptRatio:
		feedback.Conclusion = "主因はLLM側の不理解・実行失敗（AgentDefect）です。"
	default:
		feedback.Conclusion = "PromptDefectとAgentDefectの寄与は同程度、または判定材料が不足しています。"
	}
	if len(attr.PromptDefects) == 0 {
		feedback.UserFeedback = []string{"指示側の明確な欠陥は記録されていません。"}
	} else {
		feedback.UserFeedback = []string{"次回は以下の指示上の不足を明示してください: " + strings.Join(attr.PromptDefects, ", ")}
	}
	if len(attr.AgentDefects) == 0 {
		feedback.AgentCorrection = []string{"LLM側の明確な不理解は記録されていません。"}
	} else {
		feedback.AgentCorrection = []string{"LLM側は以下を再発防止対象にします: " + strings.Join(attr.AgentDefects, ", ")}
	}
	return feedback
}

func EvaluateAndInsert(ctx context.Context, pool *pgxpool.Pool, in EvaluationInput) (*MemoryRecord, error) {
	var err error
	in, err = in.normalize()
	if err != nil {
		return nil, err
	}
	history, err := ListEvaluationHistoryHeavy(ctx, pool, in.ComparisonKey, 1)
	if err != nil {
		return nil, fmt.Errorf("前回評価の取得に失敗いたしましたわ: %w", err)
	}
	var previous *EvaluationPrevious
	if len(history) > 0 {
		previous, err = previousEvaluation(&history[0])
		if err != nil {
			return nil, err
		}
	}
	meta := calculateEvaluation(in, previous)
	content, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("評価本文のJSON化に失敗いたしましたわ: %w", err)
	}
	tags := []string{"eval"}
	if in.TaskID != "" {
		tags = append(tags, "task_id:"+in.TaskID)
	}
	if in.Role != "" {
		tags = append(tags, "role:"+in.Role)
	}
	if in.TaskFamily != "" {
		tags = append(tags, "task:"+in.TaskFamily)
	}
	return InsertMemoryHeavy(ctx, pool, AddMemoryInput{
		Category:  "eval",
		Title:     evaluationTitle(in),
		ContentL0: string(content),
		Tags:      tags,
		Metadata:  structToMap(meta),
	})
}

func evaluationTitle(in EvaluationInput) string {
	if in.TaskID == "" {
		return "eval:" + in.ComparisonKey + ":" + in.EvaluationID
	}
	return "eval:task:" + in.TaskID + ":" + in.EvaluationID
}

func structToMap(value interface{}) map[string]interface{} {
	encoded, _ := json.Marshal(value)
	var result map[string]interface{}
	_ = json.Unmarshal(encoded, &result)
	return result
}

func ListEvaluationHistoryHeavy(ctx context.Context, pool *pgxpool.Pool, comparisonKey string, limit int) ([]MemoryRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	const query = `
		SELECT id, client_id, category, title, content_l0, content_l1, content_l2, tags,
		       current_level, valid_from, valid_to, version, metadata, created_at, updated_at
		FROM v_active_memories
		WHERE category = 'eval'
		  AND metadata->>'record_type' = 'eval'
		  AND metadata->>'comparison_key' = $1
		ORDER BY created_at DESC
		LIMIT $2`
	rows, err := pool.Query(ctx, query, strings.TrimSpace(comparisonKey), limit)
	if err != nil {
		return nil, fmt.Errorf("評価履歴クエリ実行エラーでしてよ: %w", err)
	}
	defer rows.Close()

	var records []MemoryRecord
	for rows.Next() {
		var rec MemoryRecord
		var rawMeta []byte
		if err := rows.Scan(&rec.ID, &rec.ClientID, &rec.Category, &rec.Title,
			&rec.ContentL0, &rec.ContentL1, &rec.ContentL2, &rec.Tags,
			&rec.CurrentLevel, &rec.ValidFrom, &rec.ValidTo, &rec.Version,
			&rawMeta, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, fmt.Errorf("評価履歴レコードのスキャンに失敗いたしましたわ: %w", err)
		}
		if err := json.Unmarshal(rawMeta, &rec.Metadata); err != nil {
			return nil, fmt.Errorf("評価履歴metadataのJSON解析に失敗いたしましたわ: %w", err)
		}
		rec.Status = "ACTIVE"
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("評価履歴走査エラーでしてよ: %w", err)
	}
	return records, nil
}
