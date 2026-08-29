// Morphism: RawDocument ∘ CompactionEngine → PipelineResult{Action, MemoryRecord, Nodes, Edges} ∘ MutatedDB
// Functor: F(IngestionStream) ⇒ Category(SelfEditingBiTemporalState)

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// IngestAction はJITMIND流自己編集の判定結果ですわ
type IngestAction string

const (
	ActionAdd       IngestAction = "ADD"
	ActionUpdate    IngestAction = "UPDATE"
	ActionNoop      IngestAction = "NOOP"
	ActionDeprecate IngestAction = "DEPRECATE"
)

// IngestResult は統合取り込み処理の結果サマリーですわ
type IngestResult struct {
	Action           IngestAction        `json:"action"`
	Reason           string              `json:"reason"`
	Memory           *MemoryRecord       `json:"memory,omitempty"`
	OldMemoryID      string              `json:"old_memory_id,omitempty"`
	Extracted        *ExtractedKnowledge `json:"extracted"`
	CreatedNodeCount int                 `json:"created_node_count"`
	CreatedEdgeCount int                 `json:"created_edge_count"`
	GraphErrors      []string            `json:"graph_errors,omitempty"`
}

// IngestKnowledgeHeavy はJITMIND自己編集判定と多段縮約・グラフ抽出を一括実行いたしますわ
func IngestKnowledgeHeavy(ctx context.Context, pool *pgxpool.Pool, title string, rawContent string, category string, forceAdd bool) (*IngestResult, error) {
	if strings.TrimSpace(rawContent) == "" {
		return nil, fmt.Errorf("取り込み対象のコンテンツが空でしてよ")
	}

	// 1. 多段縮約 & 知識グラフ抽出
	extracted, err := ExtractKnowledgeComplex(ctx, title, rawContent, category)
	if err != nil {
		return nil, fmt.Errorf("ナレッジ抽出パイプラインエラーですわ: %w", err)
	}

	if title != "" {
		extracted.Title = title
	}
	if category != "" {
		extracted.Category = category
	}

	// 2. 既存記憶との照合 (JITMIND Self-Editing Decision)
	action := ActionAdd
	reason := "新規トピックとして登録"
	var matchedOldID string

	if !forceAdd {
		existing, err := SearchMemoriesHeavy(ctx, pool, extracted.Title, "", extracted.Category, 3)
		if err == nil && len(existing) > 0 {
			for _, m := range existing {
				if normalizeIdentity(m.Title) == normalizeIdentity(extracted.Title) {
					if normalizeIdentity(m.ContentL0) == normalizeIdentity(rawContent) {
						return &IngestResult{
							Action:    ActionNoop,
							Reason:    fmt.Sprintf("同一内容の記憶 (ID: %s) が既に存在するため更新をスキップいたしましたわ", m.ID),
							Memory:    &m,
							Extracted: extracted,
						}, nil
					}

					action = ActionUpdate
					reason = fmt.Sprintf("既存記憶 (ID: %s, v%d) の更新・自己編集を検知", m.ID, m.Version)
					matchedOldID = m.ID
					break
				}
			}
		}
	}

	// 3. 判定に応じた DB 永続化 (トランザクション)
	var finalMemory *MemoryRecord
	input := AddMemoryInput{
		Category:  extracted.Category,
		Title:     extracted.Title,
		ContentL0: rawContent,
		ContentL1: extracted.ContentL1,
		ContentL2: extracted.ContentL2,
		Tags:      extracted.Tags,
		Metadata: map[string]interface{}{
			"is_llm_extracted": extracted.IsLLMUsed,
			"node_count":       len(extracted.Nodes),
			"edge_count":       len(extracted.Edges),
		},
	}

	if action == ActionUpdate && matchedOldID != "" {
		finalMemory, err = SupersedeMemoryHeavy(ctx, pool, matchedOldID, input)
		if err != nil {
			return nil, fmt.Errorf("自己編集 (UPDATE) 実行失敗ですわ: %w", err)
		}
	} else {
		finalMemory, err = InsertMemoryHeavy(ctx, pool, input)
		if err != nil {
			return nil, fmt.Errorf("新規登録 (ADD) 実行失敗ですわ: %w", err)
		}
	}

	// 4. 抽出された知識グラフノードとエッジの登録
	nodeCount := 0
	edgeCount := 0
	var graphErrors []string

	for _, n := range extracted.Nodes {
		_, err := UpsertNodeHeavy(ctx, pool, n.Name, n.NodeType, n.Summary, nil)
		if err == nil {
			nodeCount++
		} else {
			graphErrors = append(graphErrors, fmt.Sprintf("node %q: %v", n.Name, err))
		}
	}

	for _, e := range extracted.Edges {
		_, err := AddEdgeHeavy(ctx, pool, e.SourceName, e.TargetName, e.RelationType, e.Weight, finalMemory.ID)
		if err == nil {
			edgeCount++
		} else {
			graphErrors = append(graphErrors, fmt.Sprintf("edge %q -> %q: %v", e.SourceName, e.TargetName, err))
		}
	}

	return &IngestResult{
		Action:           action,
		Reason:           reason,
		Memory:           finalMemory,
		OldMemoryID:      matchedOldID,
		Extracted:        extracted,
		CreatedNodeCount: nodeCount,
		CreatedEdgeCount: edgeCount,
		GraphErrors:      graphErrors,
	}, nil
}

// IngestFileSectionsHeavy は複数セクションを含むファイルを分割して一括インジェストいたしますわ
func IngestFileSectionsHeavy(ctx context.Context, pool *pgxpool.Pool, filePath string, category string, forceAdd bool) ([]*IngestResult, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイル読み込みエラーですわ: %w", err)
	}

	rawText := string(bytes)
	if category == "" {
		category = InferCategoryFromPath(filePath)
	}

	// セクション分割正規表現 (### YYYY-MM-DD または ## [DECISION-xxx] または ## [ISSUE-xxx] または <api id=...>)
	sectionDivider := regexp.MustCompile(`(?m)^(?:###\s+\d{4}-\d{2}-\d{2}|##\s+\[DECISION-\d+\]|##\s+\[ISSUE-\d+\]|<api\s+id=)`)
	locs := sectionDivider.FindAllStringIndex(rawText, -1)

	var sections []string
	if len(locs) <= 1 {
		// 分割不要または単一セクション
		sections = append(sections, rawText)
	} else {
		for i := 0; i < len(locs); i++ {
			start := locs[i][0]
			end := len(rawText)
			if i+1 < len(locs) {
				end = locs[i+1][0]
			}
			sec := strings.TrimSpace(rawText[start:end])
			if sec != "" {
				sections = append(sections, sec)
			}
		}
	}

	var results []*IngestResult
	var sectionErrors []string
	for _, sec := range sections {
		// セクションからタイトルを推定
		lines := strings.Split(sec, "\n")
		secTitle := ""
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if strings.HasPrefix(trimmed, "#") {
				secTitle = strings.TrimLeft(trimmed, "# ")
				break
			}
		}
		if secTitle == "" {
			secTitle = fmt.Sprintf("%s entry (%s)", strings.Title(category), filepath.Base(filePath))
		}

		res, err := IngestKnowledgeHeavy(ctx, pool, secTitle, sec, category, forceAdd)
		if err != nil {
			sectionErrors = append(sectionErrors, fmt.Sprintf("%s: %v", secTitle, err))
			continue
		}
		results = append(results, res)
	}

	if len(sectionErrors) > 0 {
		return results, fmt.Errorf("%d section(s) failed: %s", len(sectionErrors), strings.Join(sectionErrors, "; "))
	}
	return results, nil
}

// BatchCompactMemoriesHeavy は未縮約の古い記憶を一括スキャンして多段要約を補完いたしますわ
func BatchCompactMemoriesHeavy(ctx context.Context, pool *pgxpool.Pool, limit int) (int, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, category, title, content_l0 
		FROM memories 
		WHERE status = 'ACTIVE' 
		  AND tx_invalidated_at IS NULL
		  AND (content_l1 IS NULL OR content_l1 = '' OR content_l2 IS NULL OR content_l2 = '')
		ORDER BY created_at ASC 
		LIMIT $1;
	`

	rows, err := pool.Query(ctx, query, limit)
	if err != nil {
		return 0, fmt.Errorf("未縮約レコードの抽出クエリ失敗ですわ: %w", err)
	}
	defer rows.Close()

	type targetItem struct {
		id       string
		category string
		title    string
		content  string
	}

	var targets []targetItem
	for rows.Next() {
		var item targetItem
		if err := rows.Scan(&item.id, &item.category, &item.title, &item.content); err == nil {
			targets = append(targets, item)
		}
	}

	compactedCount := 0
	for _, t := range targets {
		extracted, err := ExtractKnowledgeComplex(ctx, t.title, t.content, t.category)
		if err != nil {
			continue
		}

		updateQuery := `
			UPDATE memories 
			SET content_l1 = COALESCE(NULLIF(content_l1, ''), $1),
			    content_l2 = COALESCE(NULLIF(content_l2, ''), $2),
			    tags = CASE WHEN tags = '{}' OR tags IS NULL THEN $3 ELSE tags END,
			    current_level = 1,
			    updated_at = NOW()
			WHERE id = $4;
		`
		_, err = pool.Exec(ctx, updateQuery, extracted.ContentL1, extracted.ContentL2, extracted.Tags, t.id)
		if err == nil {
			compactedCount++
		}
	}

	return compactedCount, nil
}
