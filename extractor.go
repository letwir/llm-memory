// Morphism: RawText ∘ SystemConfig ∘ CategorySpec → ExtractedKnowledge{L1, L2, L3, Nodes, Edges} ∘ PurityProof
// Functor: F(MultiCategoryKnowledgeStream) ⇒ Category(StructuredTriples_Compaction)

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ExtractedKnowledge は抽出された多段要約および知識グラフトリプルですわ
type ExtractedKnowledge struct {
	Title     string          `json:"title"`
	Category  string          `json:"category"`
	ContentL1 string          `json:"content_l1"` // L1: 要点箇条書き (~30%)
	ContentL2 string          `json:"content_l2"` // L2: 1行要約 (~5%)
	Tags      []string        `json:"tags"`       // L3: タグ配列 (~1%)
	Nodes     []ExtractedNode `json:"nodes"`
	Edges     []ExtractedEdge `json:"edges"`
	IsLLMUsed bool            `json:"is_llm_used"`
	Metadata  map[string]any  `json:"metadata,omitempty"`
}

type ExtractedNode struct {
	Name     string `json:"name"`
	NodeType string `json:"node_type"` // 'technology', 'concept', 'rule', 'file', 'server', 'defect', 'decision', 'issue'
	Summary  string `json:"summary"`
}

type ExtractedEdge struct {
	SourceName   string  `json:"source_name"`
	TargetName   string  `json:"target_name"`
	RelationType string  `json:"relation_type"` // 'DEPENDS_ON', 'SUPERSEDES', 'USES', 'CONFIGURED_IN', 'ATTRIBUTED_TO', 'GOVERNS', 'TARGETS'
	Weight       float64 `json:"weight"`
}

// InferCategoryFromPath はファイル名から自動で適切なカテゴリを推定いたしますわ
func InferCategoryFromPath(filePath string) string {
	base := strings.ToLower(filepath.Base(filePath))
	switch {
	case strings.Contains(base, "diary"):
		return "diary"
	case strings.Contains(base, "decision"):
		return "decision"
	case strings.Contains(base, "method"):
		return "method"
	case strings.Contains(base, "knowledge"):
		return "knowledge"
	case strings.Contains(base, "issue"):
		return "issue"
	case strings.Contains(base, "history"):
		return "history"
	case strings.Contains(base, "walkthrough"):
		return "walkthrough"
	case strings.Contains(base, "plan") || strings.Contains(base, "implementation"):
		return "plan"
	default:
		return "knowledge"
	}
}

// ExtractKnowledgeComplex はLLMまたはヒューリスティックを用いてカテゴリ特化多段要約とグラフを抽出いたしますわ
func ExtractKnowledgeComplex(ctx context.Context, title string, rawContent string, category string) (*ExtractedKnowledge, error) {
	if strings.TrimSpace(rawContent) == "" {
		return nil, fmt.Errorf("抽出対象の本文が空でしてよ")
	}

	if category == "" {
		category = "knowledge"
	}

	apiKey := resolveGeminiAPIKey()
	if apiKey != "" {
		extracted, err := extractViaGeminiLLMHeavy(ctx, apiKey, title, rawContent, category)
		if err == nil && extracted != nil {
			extracted.IsLLMUsed = true
			if title != "" && extracted.Title == "" {
				extracted.Title = title
			}
			return extracted, nil
		}
	}

	// 決定論的ヒューリスティック抽出器
	return extractViaHeuristicsHeavy(title, rawContent, category), nil
}

func resolveGeminiAPIKey() string {
	if k := os.Getenv("GEMINI_API_KEY"); k != "" {
		return k
	}
	if k := os.Getenv("GEMINI_GROUNDING_API_KEY"); k != "" {
		return k
	}
	return ""
}

// extractViaGeminiLLMHeavy はGemini APIを呼び出してカテゴリ特化JSON構造化データを抽出いたしますわ
func extractViaGeminiLLMHeavy(ctx context.Context, apiKey string, title string, rawContent string, category string) (*ExtractedKnowledge, error) {
	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)

	categoryPromptGuide := ""
	switch category {
	case "diary":
		categoryPromptGuide = "【Diary特化指示】: 本音・愚痴、摩擦の因果判定（PromptDefect vs AgentDefectの比率や原因）、反省点を抽出し、Defectに関するノードとATTRIBUTED_TO/CAUSED_BYエッジを必ず作成してください。"
	case "decision":
		categoryPromptGuide = "【Decision特化指示】: 決定事項（DECISION-xxx）、アーキテクチャ方針、採用理由、制約を抽出し、GOVERNSエッジを作成してください。"
	case "issue":
		categoryPromptGuide = "【Issue特化指示】: Issue ID、タスク概要、状態（TODO/WIP/DONE）、ブロッカーを抽出し、TARGETSエッジを作成してください。"
	case "method":
		categoryPromptGuide = "【Method特化指示】: 設計思想、コーディング規約、CT健全性原則を抽出し、ENFORCESエッジを作成してください。"
	}

	prompt := fmt.Sprintf(`あなたはエージェント記憶基盤のナレッジグラフ抽出・多段縮約エンジンです。
以下の入力テキストから、指定のJSON形式で多段要約（L1, L2, L3）および知識グラフ（ノード、有向エッジ）を抽出してください。

【入力タイトル】: %s
【カテゴリ】: %s
%s

【入力テキスト (L0 Raw)】:
%s

【出力JSONスキーマ】:
{
  "title": "簡潔なタイトル (入力タイトルがあれば尊重)",
  "category": "%s",
  "content_l1": "・要点1\n・要点2\n・要点3 (全体の30%%程度の構造化箇条書き)",
  "content_l2": "この知識/記録の本質を一文で表現したエグゼクティブサマリー (全体の5%%)",
  "tags": ["タグ1", "タグ2", "タグ3"],
  "nodes": [
    {"name": "エンティティ名", "node_type": "technology|concept|rule|file|server|defect|decision|issue", "summary": "概要"}
  ],
  "edges": [
    {"source_name": "始点エンティティ", "target_name": "終点エンティティ", "relation_type": "DEPENDS_ON|USES|SUPERSEDES|CONFIGURED_IN|ATTRIBUTED_TO|GOVERNS|TARGETS", "weight": 1.0}
  ]
}
※ 必ずJSONのみを出力してください。`, title, category, categoryPromptGuide, rawContent, category)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseMimeType": "application/json",
			"temperature":      0.1,
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("GeminiリクエストのJSON生成に失敗いたしましたわ: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("HTTPリクエスト生成失敗ですわ: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Gemini API呼び出しに失敗いたしましたわ: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini APIエラー (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("Geminiレスポンスのデコードに失敗いたしましたわ: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("Geminiから有効な候補が返却されませんでしたわ")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var extracted ExtractedKnowledge
	if err := json.Unmarshal([]byte(responseText), &extracted); err != nil {
		return nil, fmt.Errorf("抽出JSONのUnmarshal失敗ですわ: %w (Raw: %s)", err, responseText)
	}

	return &extracted, nil
}

// extractViaHeuristicsHeavy はルールベースで決定論的にカテゴリ特化要約・タグ・グラフを抽出いたしますわ
func extractViaHeuristicsHeavy(title string, rawContent string, category string) *ExtractedKnowledge {
	lines := strings.Split(rawContent, "\n")
	var keyPoints []string
	var tags []string
	var nodes []ExtractedNode
	var edges []ExtractedEdge

	tagSet := make(map[string]bool)
	nodeSet := make(map[string]bool)

	// カテゴリ別タグ付与
	if category != "" {
		tagSet[category] = true
		tags = append(tags, category)
	}

	// 見出し・箇条書きの抽出 (L1)
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "・") {
			keyPoints = append(keyPoints, trimmed)
		} else if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "### ") {
			h := strings.TrimLeft(trimmed, "# ")
			keyPoints = append(keyPoints, fmt.Sprintf("【%s】", h))
			if !tagSet[h] && len(h) < 30 {
				tags = append(tags, strings.ToLower(h))
				tagSet[h] = true
			}
		}
	}

	// タイトルの補完
	if title == "" {
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "### ") {
				title = strings.TrimLeft(trimmed, "# ")
				break
			}
		}
		if title == "" {
			title = fmt.Sprintf("Untitled %s Entry", strings.Title(category))
		}
	}

	// L1 (要点)
	l1 := strings.Join(keyPoints, "\n")
	if l1 == "" {
		if len(lines) > 3 {
			l1 = strings.Join(lines[:3], "\n")
		} else {
			l1 = rawContent
		}
	}

	// L2 (1行要約: UTF-8ルーン安全スライス)
	var l2 string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "```") {
			runes := []rune(trimmed)
			if len(runes) > 100 {
				l2 = string(runes[:97]) + "..."
			} else {
				l2 = trimmed
			}
			break
		}
	}
	if l2 == "" {
		l2 = title
	}

	// --- カテゴリ特化パーサー ---
	switch category {
	case "diary":
		// 因果判定比率 [PromptDefect: xx%] vs [AgentDefect: xx%]
		ratioRe := regexp.MustCompile(`\[(?:ワイの指示|PromptDefect)[\(:]?\s*([^\]]+)\]\s*vs\s*\[(?:AIの認知|AgentDefect)[\(:]?\s*([^\]]+)\]`)
		if matches := ratioRe.FindStringSubmatch(rawContent); len(matches) >= 3 {
			tags = append(tags, "friction-ratio")
			l2 = fmt.Sprintf("[因果比率: %s vs %s] %s", matches[1], matches[2], l2)
		}

		// 欠陥分類キーワード抽出
		defectKeywords := []string{"PromptDefect", "AgentDefect", "Underspecified", "Contradiction", "ImplicitContext", "AmbiguousScope", "Misreading", "KnowledgeGap", "Overthinking", "ToolFailure"}
		for _, dk := range defectKeywords {
			if strings.Contains(rawContent, dk) {
				tagLower := strings.ToLower(dk)
				if !tagSet[tagLower] {
					tags = append(tags, tagLower)
					tagSet[tagLower] = true
				}
				if !nodeSet[dk] {
					nodes = append(nodes, ExtractedNode{
						Name:     dk,
						NodeType: "defect",
						Summary:  fmt.Sprintf("Defect mentioned in %s", title),
					})
					nodeSet[dk] = true
				}
			}
		}

		if len(nodes) > 0 {
			edges = append(edges, ExtractedEdge{
				SourceName:   title,
				TargetName:   nodes[0].Name,
				RelationType: "ATTRIBUTED_TO",
				Weight:       1.0,
			})
		}

	case "decision":
		// DECISION-xxx 抽出
		decRe := regexp.MustCompile(`\[DECISION-(\d+)\]`)
		for _, m := range decRe.FindAllString(rawContent, -1) {
			if !tagSet[m] {
				tags = append(tags, m)
				tagSet[m] = true
			}
			if !nodeSet[m] {
				nodes = append(nodes, ExtractedNode{
					Name:     m,
					NodeType: "decision",
					Summary:  fmt.Sprintf("Decision record: %s", title),
				})
				nodeSet[m] = true
			}
		}

	case "issue":
		// ISSUE-xxx 抽出
		issRe := regexp.MustCompile(`ISSUE-(\d+)`)
		for _, m := range issRe.FindAllString(rawContent, -1) {
			if !tagSet[m] {
				tags = append(tags, m)
				tagSet[m] = true
			}
			if !nodeSet[m] {
				nodes = append(nodes, ExtractedNode{
					Name:     m,
					NodeType: "issue",
					Summary:  fmt.Sprintf("Issue item: %s", title),
				})
				nodeSet[m] = true
			}
		}
	}

	// 汎用技術キーワード抽出
	keywords := []string{"PostgreSQL", "Postgres", "Go", "Golang", "Python", "Rust", "Docker", "Tailscale", "Debian", "Windows", "Bi-Temporal", "JITMIND", "Graphiti", "Memtrace", "Compaction", "MCP", "Zed", "Antigravity"}
	for _, kw := range keywords {
		matched, _ := regexp.MatchString("(?i)\\b"+regexp.QuoteMeta(kw)+"\\b", rawContent)
		if matched {
			tagLower := strings.ToLower(kw)
			if !tagSet[tagLower] {
				tags = append(tags, tagLower)
				tagSet[tagLower] = true
			}
			if !nodeSet[kw] {
				nodes = append(nodes, ExtractedNode{
					Name:     kw,
					NodeType: "technology",
					Summary:  fmt.Sprintf("Mentioned in %s", title),
				})
				nodeSet[kw] = true
			}
		}
	}

	// ノード共起関係性エッジ
	if len(nodes) >= 2 && len(edges) == 0 {
		for i := 0; i < len(nodes)-1; i++ {
			edges = append(edges, ExtractedEdge{
				SourceName:   nodes[i].Name,
				TargetName:   nodes[i+1].Name,
				RelationType: "ASSOCIATED_WITH",
				Weight:       1.0,
			})
		}
	}

	return &ExtractedKnowledge{
		Title:     title,
		Category:  category,
		ContentL1: l1,
		ContentL2: l2,
		Tags:      tags,
		Nodes:     nodes,
		Edges:     edges,
		IsLLMUsed: false,
	}
}
