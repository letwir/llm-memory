// Morphism: DiaryTextEntries ∘ CausalTaxonomy → AttributionStatistics ∘ FeedbackReport
// Functor: F(DiaryParsing) ⇒ Category(SelfEvolvingAttribution)
// Semantics: 因果帰属（PromptDefect vs AgentDefect）を多段抽出し、人間とAIの双方向自己進化レポートを合成する射ですわ

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

// DiaryEntry は diary.md の1回分の思考・反省・因果帰属レコードですわ
type DiaryEntry struct {
	Timestamp      string   `json:"timestamp"`
	Hypothesis     string   `json:"hypothesis"`
	Tried          string   `json:"tried"`
	Rejected       string   `json:"rejected"`
	Uncertainty    string   `json:"uncertainty"`
	AttributionRaw string   `json:"attribution_raw"`
	PromptDefects  []string `json:"prompt_defects"`
	AgentDefects   []string `json:"agent_defects"`
	PromptRatio    int      `json:"prompt_ratio"`
	AgentRatio     int      `json:"agent_ratio"`
	SearchQuery    string   `json:"search_query"`
	Correction     string   `json:"correction"`
	RawContent     string   `json:"raw_content"`
}

// AttributionAnalysisResult は因果帰属の集計結果と双方向フィードバックですわ
type AttributionAnalysisResult struct {
	TotalEntries       int            `json:"total_entries"`
	EntriesWithAttr    int            `json:"entries_with_attr"`
	TotalPromptDefects int            `json:"total_prompt_defects"`
	TotalAgentDefects  int            `json:"total_agent_defects"`
	PromptDefectFreq   map[string]int `json:"prompt_defect_freq"`
	AgentDefectFreq    map[string]int `json:"agent_defect_freq"`
	AveragePromptRatio float64        `json:"average_prompt_ratio"`
	AverageAgentRatio  float64        `json:"average_agent_ratio"`
	UserFeedbackTips   []string       `json:"user_feedback_tips"`
	RecentImprovements []string       `json:"recent_improvements"`
	AgentRuleDiffs     []string       `json:"agent_rule_diffs"`
}

var (
	headerRegex         = regexp.MustCompile(`^###\s+(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`)
	promptPctRegex      = regexp.MustCompile(`(?i)(?:PromptDefect|ワイの指示)[^\d]*(\d+)\s*%`)
	agentPctRegex       = regexp.MustCompile(`(?i)(?:AgentDefect|AIの?認知)[^\d]*(\d+)\s*%`)
	braceGroupRegex     = regexp.MustCompile(`\{([^}]+)\}`)
	knownPromptKeywords = []string{"Underspecified", "Contradiction", "ImplicitContext", "AmbiguousScope", "MissingSpec"}
	knownAgentKeywords  = []string{"Misreading", "KnowledgeGap", "Overthinking", "ToolFailure", "ConfirmationBias"}
)

// ParseDiaryEntriesHeavy はファイルまたはReaderから全日記エントリを抽出・構造化する関数ですわ
func ParseDiaryEntriesHeavy(r io.Reader) ([]DiaryEntry, error) {
	scanner := bufio.NewScanner(r)
	// Why: 巨大な反省ログでもバッファ溢れを起こさないよう1MB確保しておりますの
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var entries []DiaryEntry
	var current *DiaryEntry
	var currentLines []string

	flushCurrent := func() {
		if current == nil {
			return
		}
		current.RawContent = strings.Join(currentLines, "\n")
		extractEntryFieldsHeavy(current)
		entries = append(entries, *current)
		current = nil
		currentLines = nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		matches := headerRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			flushCurrent()
			current = &DiaryEntry{
				Timestamp: matches[1],
			}
			currentLines = append(currentLines, line)
			continue
		}

		if current != nil {
			currentLines = append(currentLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("diaryスキャン中にエラーが発生いたしましたわ: %w", err)
	}

	flushCurrent()
	return entries, nil
}

// extractEntryFieldsHeavy は1つのエントリテキストから各フィールド・因果帰属を精査・抽出しますわ
func extractEntryFieldsHeavy(e *DiaryEntry) {
	text := e.RawContent

	// Why: スラッシュ区切りまたはコロン区切りのキーバリュー行をパースするため正規化しますの
	parts := strings.Split(text, "/")
	if len(parts) == 1 {
		parts = strings.Split(text, "\n")
	}

	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "hypothesis:") {
			e.Hypothesis = strings.TrimSpace(trimmed[len("hypothesis:"):])
		} else if strings.HasPrefix(lower, "tried:") {
			e.Tried = strings.TrimSpace(trimmed[len("tried:"):])
		} else if strings.HasPrefix(lower, "rejected:") {
			e.Rejected = strings.TrimSpace(trimmed[len("rejected:"):])
		} else if strings.HasPrefix(lower, "uncertainty:") {
			e.Uncertainty = strings.TrimSpace(trimmed[len("uncertainty:"):])
		} else if strings.HasPrefix(lower, "search:") {
			e.SearchQuery = strings.TrimSpace(trimmed[len("search:"):])
		} else if strings.HasPrefix(lower, "correction:") {
			e.Correction = strings.TrimSpace(trimmed[len("correction:"):])
		} else if strings.Contains(lower, "attribution") || strings.Contains(lower, "promptdefect") || strings.Contains(lower, "agentdefect") {
			e.AttributionRaw = trimmed
		}
	}

	// 因果帰属比率の抽出 (例: [ワイの指示(PromptDefect): 65%] vs [AI認知(AgentDefect): 35%])
	if pMatches := promptPctRegex.FindStringSubmatch(text); len(pMatches) >= 2 {
		fmt.Sscanf(pMatches[1], "%d", &e.PromptRatio)
	}
	if aMatches := agentPctRegex.FindStringSubmatch(text); len(aMatches) >= 2 {
		fmt.Sscanf(aMatches[1], "%d", &e.AgentRatio)
	}

	// 波括弧群 {Underspecified(...), Contradiction(...)} または既知キーワードのスキャン
	textLower := strings.ToLower(text)

	// PromptDefects の検出
	for _, kw := range knownPromptKeywords {
		if strings.Contains(textLower, strings.ToLower(kw)) {
			e.PromptDefects = append(e.PromptDefects, kw)
		}
	}

	// AgentDefects の検出
	for _, kw := range knownAgentKeywords {
		if strings.Contains(textLower, strings.ToLower(kw)) {
			e.AgentDefects = append(e.AgentDefects, kw)
		}
	}

	// 波括弧内の要素抽出フォールバック
	if len(e.PromptDefects) == 0 && len(e.AgentDefects) == 0 {
		braceMatches := braceGroupRegex.FindAllStringSubmatch(text, -1)
		for _, bm := range braceMatches {
			items := strings.Split(bm[1], ",")
			for _, item := range items {
				clean := strings.TrimSpace(item)
				if idx := strings.Index(clean, "("); idx != -1 {
					clean = clean[:idx]
				}
				if clean != "" {
					e.PromptDefects = append(e.PromptDefects, clean)
				}
			}
		}
	}
}

// AnalyzeAttributionEntriesHeavy は抽出された全日記から因果帰属統計とフィードバックを計算・集計しますわ
func AnalyzeAttributionEntriesHeavy(entries []DiaryEntry) AttributionAnalysisResult {
	res := AttributionAnalysisResult{
		TotalEntries:     len(entries),
		PromptDefectFreq: make(map[string]int),
		AgentDefectFreq:  make(map[string]int),
	}

	var sumPromptRatio, sumAgentRatio float64
	var ratioCount int

	for _, e := range entries {
		hasAttr := false

		if len(e.PromptDefects) > 0 || len(e.AgentDefects) > 0 || e.PromptRatio > 0 || e.AgentRatio > 0 || e.AttributionRaw != "" {
			hasAttr = true
			res.EntriesWithAttr++
		}

		for _, pd := range e.PromptDefects {
			res.PromptDefectFreq[pd]++
			res.TotalPromptDefects++
		}

		for _, ad := range e.AgentDefects {
			res.AgentDefectFreq[ad]++
			res.TotalAgentDefects++
		}

		if e.PromptRatio > 0 || e.AgentRatio > 0 {
			sumPromptRatio += float64(e.PromptRatio)
			sumAgentRatio += float64(e.AgentRatio)
			ratioCount++
		} else if hasAttr {
			// デフォルト比率の推定
			pCount := len(e.PromptDefects)
			aCount := len(e.AgentDefects)
			if pCount+aCount > 0 {
				pRatio := float64(pCount) / float64(pCount+aCount) * 100.0
				sumPromptRatio += pRatio
				sumAgentRatio += (100.0 - pRatio)
				ratioCount++
			}
		}
	}

	if ratioCount > 0 {
		res.AveragePromptRatio = sumPromptRatio / float64(ratioCount)
		res.AverageAgentRatio = sumAgentRatio / float64(ratioCount)
	}

	// 旦那様向けフィードバック (PromptDefect からの教訓)
	res.UserFeedbackTips = generateUserFeedbackTips(res.PromptDefectFreq)
	res.RecentImprovements = generateRecentImprovements(entries)

	// AIエージェント向けルール自己進化 Diff (AgentDefect からのルールパッチ)
	res.AgentRuleDiffs = generateAgentRuleDiffs(res.AgentDefectFreq)

	return res
}

// generateRecentImprovements は直近の記録から、次回の指示にそのまま転用できる改善文を作成しますわ。
// 頻度集計の一般論とは分離し、diaryの新しいエントリ順を優先して重複を除去いたしますの。
func generateRecentImprovements(entries []DiaryEntry) []string {
	var improvements []string
	seen := make(map[string]bool)

	for i := len(entries) - 1; i >= 0; i-- {
		for _, defect := range entries[i].PromptDefects {
			key := strings.ToLower(defect)
			if seen[key] {
				continue
			}
			seen[key] = true

			var action string
			switch {
			case strings.Contains(key, "underspec"), strings.Contains(key, "ambiguous"), strings.Contains(key, "missing"):
				action = "次回は対象範囲・入力形式・出力形式・完了条件を指示に明記する。"
			case strings.Contains(key, "implicit"), strings.Contains(key, "context"):
				action = "次回は参照ファイル、既存決定事項、実行環境などの前提コンテキストを指示に添える。"
			case strings.Contains(key, "contradict"):
				action = "次回は過去の決定事項との整合性確認と、矛盾時の優先順位を指示に明記する。"
			default:
				action = fmt.Sprintf("次回は「%s」を避けるための制約と受入条件を指示冒頭に明記する。", defect)
			}
			improvements = append(improvements, action)
		}
	}

	return improvements
}

// generateUserFeedbackTips は PromptDefect の傾向から旦那様への建設的プロンプト改善提案を合成しますわ
func generateUserFeedbackTips(freq map[string]int) []string {
	var tips []string

	if len(freq) == 0 {
		tips = append(tips, "直近の指示において重大な PromptDefect（指示側の欠陥）は検知されませんでした。極めて明瞭な指示を賜り感謝申し上げますわ！")
		return tips
	}

	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range freq {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })

	for _, item := range sorted {
		kLower := strings.ToLower(item.Key)
		switch {
		case strings.Contains(kLower, "underspec") || strings.Contains(kLower, "ambiguous"):
			tips = append(tips, fmt.Sprintf("【指示の具体化推奨 (%d回検知)】: 入力フォーマットや出力形式（CLI引数や戻り値の型）を1行追記いただくだけで、推測による試行錯誤を大幅に削減できますわ。", item.Value))
		case strings.Contains(kLower, "implicit") || strings.Contains(kLower, "context"):
			tips = append(tips, fmt.Sprintf("【前提コンテキストの明示 (%d回検知)】: プロジェクト固有の命名規約やディレクトリ配置の前提を `@file` 等で指示に添えていただくと、初回での適合率が跳ね上がりますの。", item.Value))
		case strings.Contains(kLower, "contradict"):
			tips = append(tips, fmt.Sprintf("【要件の整合性チェック (%d回検知)】: 過去の決定事項との整合性（decisions.md）に留意した指示をいただけると、差し戻しを防げますわ。", item.Value))
		default:
			tips = append(tips, fmt.Sprintf("【%s 要因への対策 (%d回検知)】: 該当する制約条件をプロンプト冒頭に箇条書きでご指定いただくと効果的ですわ。", item.Key, item.Value))
		}
	}

	return tips
}

// generateAgentRuleDiffs は AgentDefect の傾向から PERSONA.css への自己修正パッチを合成しますわ
func generateAgentRuleDiffs(freq map[string]int) []string {
	var diffs []string

	if len(freq) == 0 {
		return diffs
	}

	for k, v := range freq {
		kLower := strings.ToLower(k)
		switch {
		case strings.Contains(kLower, "tool") || strings.Contains(kLower, "command"):
			diffs = append(diffs, fmt.Sprintf(`/* 提案: ツール誤用防止 (%d回検知) */
[domain~="CLI"] {
  guard: "実行前に引数構文 (--help) の静的検証を必須化";
}`, v))
		case strings.Contains(kLower, "misreading") || strings.Contains(kLower, "overthinking"):
			diffs = append(diffs, fmt.Sprintf(`/* 提案: 早合点・過剰推論の抑制 (%d回検知) */
.agent {
  axiom: "指示内容の自己補完を禁止し、未定義パラメータは早期リターンで確認";
}`, v))
		case strings.Contains(kLower, "knowledge") || strings.Contains(kLower, "gap"):
			diffs = append(diffs, fmt.Sprintf(`/* 提案: ドメイン知識ギャップ対策 (%d回検知) */
.agent::research {
  priority: "LLM内生知識に依存せず、常に公式リファレンス (ETLパイプライン) を先読み";
}`, v))
		}
	}

	return diffs
}

// AnalyzeDiaryFileHeavy は指定パスの diary.md をパースしてコンソール出力を行うCLIハンドラですわ
func AnalyzeDiaryFileHeavy(ctx context.Context, filePath string, asJSON bool, suggestDiff bool) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("diaryファイルを開けませんでしたわ (%s): %w", filePath, err)
	}
	defer f.Close()

	entries, err := ParseDiaryEntriesHeavy(f)
	if err != nil {
		return err
	}

	result := AnalyzeAttributionEntriesHeavy(entries)

	if asJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	fmt.Println("================================================================================")
	fmt.Println("📊 因果帰属（Causal Attribution）＆ 双方向自己進化分析レポート")
	fmt.Println("================================================================================")
	fmt.Printf("総エントリ数          : %d 件\n", result.TotalEntries)
	fmt.Printf("因果帰属記録済み      : %d 件\n", result.EntriesWithAttr)
	fmt.Printf("PromptDefect (指示側) : %d 件 (平均寄与率: %.1f%%)\n", result.TotalPromptDefects, result.AveragePromptRatio)
	fmt.Printf("AgentDefect  (AI側)   : %d 件 (平均寄与率: %.1f%%)\n", result.TotalAgentDefects, result.AverageAgentRatio)
	fmt.Println("--------------------------------------------------------------------------------")

	fmt.Println("\n【1. 旦那様へのプロンプト改善フィードバック (直近の改善内容)】")
	if len(result.RecentImprovements) == 0 {
		fmt.Println(" 直近のPromptDefectから改善文は生成されませんでしたわ。")
	} else {
		for i, improvement := range result.RecentImprovements {
			fmt.Printf(" %d. %s\n", i+1, improvement)
		}
	}

	fmt.Println("\n【1.5 PromptDefectの頻度傾向（補足）】")
	for i, tip := range result.UserFeedbackTips {
		fmt.Printf(" %d. %s\n", i+1, tip)
	}

	fmt.Println("\n【2. AIエージェント自身の反省・自己ルール最適化 (AgentDefect 分析)】")
	if len(result.AgentRuleDiffs) == 0 {
		fmt.Println(" 特筆すべき反復AgentDefectは検出されませんでしたわ。")
	} else {
		for i, diff := range result.AgentRuleDiffs {
			fmt.Printf(" [最適化パッチ候補 %d]\n%s\n\n", i+1, diff)
		}
	}

	if suggestDiff && len(result.AgentRuleDiffs) > 0 {
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println("💡 ヒント: 上記パッチを PERSONA.css や MACHINE.toml に適用することで、次回以降の同種ミスを自動抑止できますわ。")
	}

	return nil
}
