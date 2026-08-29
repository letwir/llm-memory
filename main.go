// Morphism: CLIArguments ∘ SystemContext → ExecutedAction ∘ StandardOutput
// Functor: F(CLIInvocation) ⇒ Category(AgentMemoryEffect)

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func splitCommaList(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

// GlobalCLITimeout はCLIコマンド全体の最大実行制限時間ですわ
const GlobalCLITimeout = 45 * time.Second

func printUsage() {
	fmt.Println(`=== LLM Memory CLI (Bi-Temporal & Multi-Level Knowledge Base) ===
Usage:
  llm-mem status                  : DB疎通確認・端末自動登録・統計情報
  llm-mem ingest [options]        : 統合ナレッジ取り込み (JITMIND自己編集 + 自動多段縮約 + グラフ抽出)
  llm-mem compact [options]       : 未縮約の古い記憶のバッチ多段要約処理
  llm-mem add [options]           : 新規記憶の登録 (L0~L3多段対応)
  llm-mem eval [options]          : 評価を計算して新規イベントとして登録
  llm-mem evals [options]         : 評価履歴の参照
  llm-mem stock [options]         : ストック記憶の簡易一覧表示
  llm-mem search [options]        : 記憶の横断検索 (キーワード/タグ/カテゴリ)
  llm-mem supersede [options]     : 古い記憶の無効化と後続新記憶の登録 (JITMIND流)
  llm-mem analyze [options]       : diary.md 因果帰属分析 & 双方向フィードバック生成
  llm-mem graph node [options]    : 知識グラフノードの登録/更新
  llm-mem graph edge [options]    : 知識グラフ関係性エッジの登録
  llm-mem graph list [options]    : 有効な知識グラフ関係性の一覧表示
  llm-mem clients                 : 登録済みクライアント端末一覧

Options for 'analyze':
  -file     <string>  : 解析する diary.md ファイルパス (default: 'diary.md' または自動検出)
  -json               : JSON形式で構造化出力
  -suggest            : PERSONA.css / MACHINE.toml 最適化パッチDiffの表示

Options for 'ingest' (Recommended):
  -file     <string>  : 取り込むMarkdown/テキストファイルのパス
  -text     <string>  : 直接取り込むテキスト本文
  -title    <string>  : タイトル (省略時はファイル名や1行目から自動推定)
  -cat      <string>  : カテゴリ (default: 'knowledge', 'walkthrough', 'rule')
  -force              : 既存記憶の自己編集判定をスキップし強制新規登録

Options for 'compact':
  -limit    <int>     : バッチ処理する未縮約レコード上限数 (default: 20)

Options for 'add':
  -title    <string>  : 記憶のタイトル (必須)
  -content  <string>  : L0 Raw 本文 (必須)
  -cat      <string>  : カテゴリ (default: 'knowledge')
  -l1       <string>  : L1 要点箇条書き (~30%要約)
  -l2       <string>  : L2 1行要約 (~5%要約)
  -tags     <string>  : カンマ区切りタグ
  -type     <string>  : Memory Object種別 (invariant/decision/constraint/failure/knowledge/procedure/state)
  -scope    <string>  : 適用範囲 (global/project/subsystem:<name>)
  -rationale <string> : 採用理由
  -evidence <string>  : 根拠識別子のカンマ区切り
  -rejected <string>  : 却下案のカンマ区切り
  -confidence <float> : 確信度 (0..1)

Options for 'eval':
  -file     <string>  : 評価入力JSONファイル (必須)
  -json               : 登録結果をJSON形式で出力

Evaluation input additions:
  task_id              : タスク識別子（タスク終了評価では必須）
  attribution          : PromptDefect / AgentDefect と寄与比率

Options for 'evals':
  -key      <string>  : comparison_key (必須)
  -limit    <int>     : 取得件数 (default: 20)
  -json               : 評価履歴をJSON形式で出力

Options for 'stock':
  -cat      <string>  : カテゴリフィルタ
  -limit    <int>     : 取得件数 (default: 20)
  -json               : 記憶レコードをJSON形式で出力

Options for 'search':
  -q        <string>  : 検索キーワード (タイトル・本文部分一致)
  -tag      <string>  : 指定タグ
  -cat      <string>  : カテゴリフィルタ
  -level    <int>     : 表示詳細度 (0:Raw, 1:L1, 2:L2, 3:TagsOnly, default: 1)
  -limit    <int>     : 取得件数 (default: 10)
  -json               : JSON形式で出力

Options for 'supersede':
  -id       <string>  : 無効化する古い記憶のUUID (必須)
  -title    <string>  : 新しい記憶のタイトル (必須)
  -content  <string>  : 新しいL0 Raw 本文 (必須)
  -l1       <string>  : 新しいL1 要点
  -l2       <string>  : 新しいL2 1行要約
  -tags     <string>  : カンマ区切りタグ`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalCLITimeout)
	defer cancel()

	cmd := os.Args[1]

	switch cmd {
	case "status", "ping":
		handleStatusHeavy(ctx)
	case "ingest":
		handleIngestHeavy(ctx, os.Args[2:])
	case "compact":
		handleCompactHeavy(ctx, os.Args[2:])
	case "add":
		handleAddHeavy(ctx, os.Args[2:])
	case "eval":
		handleEvalHeavy(ctx, os.Args[2:])
	case "evals":
		handleEvalsHeavy(ctx, os.Args[2:])
	case "stock":
		handleStockHeavy(ctx, os.Args[2:])
	case "search":
		handleSearchHeavy(ctx, os.Args[2:])
	case "supersede", "update":
		handleSupersedeHeavy(ctx, os.Args[2:])
	case "analyze", "evolve":
		handleAnalyzeHeavy(ctx, os.Args[2:])
	case "graph":
		handleGraphHeavy(ctx, os.Args[2:])
	case "clients":
		handleClientsHeavy(ctx)
	default:
		fmt.Printf("未知のコマンドですわ: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

// handleEvalHeavy はLLM抽出を経由せず、評価イベントを計算して直接登録いたしますわ
func handleEvalHeavy(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	filePath := fs.String("file", "", "Evaluation input JSON file")
	asJSON := fs.Bool("json", false, "Output JSON")

	if err := fs.Parse(args); err != nil {
		fmt.Printf("引数パースエラーですわ: %v\n", err)
		os.Exit(1)
	}
	if *filePath == "" {
		fmt.Println("エラー: -file は必須指定でしてよ")
		os.Exit(1)
	}

	raw, err := os.ReadFile(*filePath)
	if err != nil {
		fmt.Printf("評価入力ファイルの読み込み失敗: %v\n", err)
		os.Exit(1)
	}
	var input EvaluationInput
	if err := json.Unmarshal(raw, &input); err != nil {
		fmt.Printf("評価入力JSONの解析失敗: %v\n", err)
		os.Exit(1)
	}

	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		fmt.Printf("DB接続失敗: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	rec, err := EvaluateAndInsert(ctx, pool, input)
	if err != nil {
		fmt.Printf("評価登録失敗: %v\n", err)
		os.Exit(1)
	}
	if *asJSON {
		out, _ := json.MarshalIndent(rec, "", "  ")
		fmt.Println(string(out))
		return
	}
	fmt.Printf("評価を登録いたしましたわ\nID: %s\nKey: %s\n", rec.ID, rec.Metadata["comparison_key"])
	if taskID, ok := rec.Metadata["task_id"].(string); ok && taskID != "" {
		fmt.Printf("Task: %s\n", taskID)
	}
	if relative, ok := rec.Metadata["relative"].(map[string]interface{}); ok {
		fmt.Printf("Relative: %v\n", relative)
	}
	if feedback, ok := rec.Metadata["feedback"].(map[string]interface{}); ok {
		fmt.Printf("\n=== タスク評価リザルト ===\n")
		if conclusion, ok := feedback["conclusion"].(string); ok {
			fmt.Printf("結論: %s\n", conclusion)
		}
		if items, ok := feedback["user_feedback"].([]interface{}); ok {
			fmt.Println("人間の指示へのフィードバック:")
			for _, item := range items {
				fmt.Printf("- %v\n", item)
			}
		}
		if items, ok := feedback["agent_correction"].([]interface{}); ok {
			fmt.Println("LLM側の不理解へのフィードバック:")
			for _, item := range items {
				fmt.Printf("- %v\n", item)
			}
		}
	}
}

// handleEvalsHeavy は評価専用の履歴参照を処理いたしますわ
func handleEvalsHeavy(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("evals", flag.ExitOnError)
	key := fs.String("key", "", "Comparison key")
	limit := fs.Int("limit", 20, "Limit")
	asJSON := fs.Bool("json", false, "Output JSON")

	if err := fs.Parse(args); err != nil {
		fmt.Printf("引数パースエラーですわ: %v\n", err)
		os.Exit(1)
	}
	if strings.TrimSpace(*key) == "" {
		fmt.Println("エラー: -key は必須指定でしてよ")
		os.Exit(1)
	}

	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		fmt.Printf("DB接続失敗: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	records, err := ListEvaluationHistoryHeavy(ctx, pool, *key, *limit)
	if err != nil {
		fmt.Printf("評価履歴取得失敗: %v\n", err)
		os.Exit(1)
	}
	if *asJSON {
		out, _ := json.MarshalIndent(records, "", "  ")
		fmt.Println(string(out))
		return
	}
	fmt.Printf("=== 評価履歴 (%d件) ===\n", len(records))
	for i, rec := range records {
		fmt.Printf("[%d] %s | ID: %s | Created: %s\n", i+1, rec.Title, rec.ID, rec.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("    Axes: %v\n", rec.Metadata["axes"])
		fmt.Printf("    Relative: %v\n", rec.Metadata["relative"])
	}
}

// handleStockHeavy は日常的にストック記憶を眺めるための簡易表示ですわ
func handleStockHeavy(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("stock", flag.ExitOnError)
	cat := fs.String("cat", "", "Category filter")
	limit := fs.Int("limit", 20, "Limit")
	asJSON := fs.Bool("json", false, "Output JSON")

	if err := fs.Parse(args); err != nil {
		fmt.Printf("引数パースエラーですわ: %v\n", err)
		os.Exit(1)
	}

	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		fmt.Printf("DB接続失敗: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	records, err := SearchMemoriesHeavy(ctx, pool, "", "", *cat, *limit)
	if err != nil {
		fmt.Printf("ストック取得失敗: %v\n", err)
		os.Exit(1)
	}
	if *asJSON {
		out, _ := json.MarshalIndent(records, "", "  ")
		fmt.Println(string(out))
		return
	}
	if len(records) == 0 {
		fmt.Println("有効なストック記憶はありませんわ。")
		return
	}
	fmt.Printf("=== ストック記憶 (%d件) ===\n", len(records))
	for i, rec := range records {
		oneLine := rec.ContentL0
		if rec.ContentL2 != nil && strings.TrimSpace(*rec.ContentL2) != "" {
			oneLine = *rec.ContentL2
		} else if rec.ContentL1 != nil && strings.TrimSpace(*rec.ContentL1) != "" {
			oneLine = *rec.ContentL1
		}
		oneLine = strings.Join(strings.Fields(oneLine), " ")
		fmt.Printf("[%02d] %-16s %-24s %s\n", i+1, rec.Category, rec.Title, oneLine)
	}
}

// handleIngestHeavy はJITMIND自己編集と多段縮約・グラフ抽出を一括実行いたしますわ
func handleIngestHeavy(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("ingest", flag.ExitOnError)
	filePath := fs.String("file", "", "Target file path to ingest")
	text := fs.String("text", "", "Direct text content")
	title := fs.String("title", "", "Title override")
	cat := fs.String("cat", "", "Category (auto-inferred from filename if empty)")
	force := fs.Bool("force", false, "Force add without checking conflicts")

	if err := fs.Parse(args); err != nil {
		fmt.Printf("引数パースエラーですわ: %v\n", err)
		os.Exit(1)
	}

	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		fmt.Printf("DB接続失敗: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if *filePath != "" {
		if *cat == "" {
			*cat = InferCategoryFromPath(*filePath)
		}
		fmt.Printf("🚀 ファイル [%s] をカテゴリ [%s] として解析・取り込み中...\n", *filePath, *cat)
		results, err := IngestFileSectionsHeavy(ctx, pool, *filePath, *cat, *force)
		if err != nil {
			fmt.Printf("ファイル取り込み失敗: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n=== ファイル一括取り込み完了サマリー (%d セクション) ===\n", len(results))
		for i, res := range results {
			fmt.Printf("[%d] 判定: [%s] (%s) | タイトル: %s\n", i+1, res.Action, res.Reason, res.Extracted.Title)
			if res.Memory != nil {
				fmt.Printf("    ID: %s (v%d) | L2: %s\n", res.Memory.ID, res.Memory.Version, *res.Memory.ContentL2)
			}
			fmt.Printf("    ノード: %d 件, エッジ: %d 件\n", res.CreatedNodeCount, res.CreatedEdgeCount)
		}
		return
	}

	if *text != "" {
		if *cat == "" {
			*cat = "knowledge"
		}
		if *title == "" {
			*title = "Direct Ingest Entry"
		}
		fmt.Printf("🚀 直接テキストをカテゴリ [%s] として解析・取り込み中...\n", *cat)
		res, err := IngestKnowledgeHeavy(ctx, pool, *title, *text, *cat, *force)
		if err != nil {
			fmt.Printf("テキスト取り込み失敗: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n=== 取り込み完了サマリー ===\n")
		fmt.Printf("・判定アクション: [%s] (%s)\n", res.Action, res.Reason)
		if res.Memory != nil {
			fmt.Printf("・記憶ID        : %s (v%d, %s)\n", res.Memory.ID, res.Memory.Version, res.Memory.Status)
			fmt.Printf("・タイトル      : %s\n", res.Memory.Title)
			fmt.Printf("・L2 要約       : %s\n", *res.Memory.ContentL2)
			fmt.Printf("・タグ          : %v\n", res.Memory.Tags)
		}
		fmt.Printf("・抽出ノード数  : %d 件\n", res.CreatedNodeCount)
		fmt.Printf("・抽出エッジ数  : %d 件\n", res.CreatedEdgeCount)
		return
	}

	fmt.Println("エラー: -file または -text のいずれかを指定してくださいまし")
	os.Exit(1)
}

// handleCompactHeavy は未縮約レコードのバッチ縮約処理を遂行いたしますわ
func handleCompactHeavy(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("compact", flag.ExitOnError)
	limit := fs.Int("limit", 20, "Batch limit")
	_ = fs.Parse(args)

	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		fmt.Printf("DB接続失敗: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	fmt.Printf("🧹 未縮約レコードのバッチ多段要約を開始いたします (上限: %d 件)...\n", *limit)
	count, err := BatchCompactMemoriesHeavy(ctx, pool, *limit)
	if err != nil {
		fmt.Printf("バッチ縮約失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✨ %d 件の記憶を多段縮約（L1/L2/L3付与）いたしましたわ！\n", count)
}

// handleStatusHeavy はDB疎通と基本統計情報を出力いたしますわ
func handleStatusHeavy(ctx context.Context) {
	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		fmt.Printf("❌ DB接続エラー: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	client, err := AutoRegisterClientHeavy(ctx, pool)
	if err != nil {
		fmt.Printf("⚠️ クライアント自動登録警告: %v\n", err)
	} else {
		fmt.Printf("🟢 端末認証成功: %s (%s, %s)\n", client.ClientID, client.Hostname, client.OSInfo)
	}

	var activeMemCount, nodeCount, edgeCount int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM v_active_memories").Scan(&activeMemCount)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM knowledge_nodes").Scan(&nodeCount)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM v_active_knowledge_graph").Scan(&edgeCount)

	fmt.Println("=== PostgreSQL LLM Memory Status ===")
	fmt.Printf("・有効アクティブ記憶数 (Active Memories): %d 件\n", activeMemCount)
	fmt.Printf("・知識グラフノード数   (Knowledge Nodes): %d 件\n", nodeCount)
	fmt.Printf("・有効グラフ関係性数   (Active Edges)   : %d 件\n", edgeCount)
}

// handleAddHeavy は新規記憶の登録処理を遂行いたしますわ
func handleAddHeavy(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	title := fs.String("title", "", "Title")
	content := fs.String("content", "", "L0 Raw Content")
	cat := fs.String("cat", "knowledge", "Category")
	l1 := fs.String("l1", "", "L1 Key Points")
	l2 := fs.String("l2", "", "L2 One-liner")
	tagsStr := fs.String("tags", "", "Comma-separated tags")
	objectType := fs.String("type", "", "Memory Object type")
	scope := fs.String("scope", "", "Memory Object scope")
	rationale := fs.String("rationale", "", "Memory Object rationale")
	evidenceStr := fs.String("evidence", "", "Comma-separated evidence references")
	rejectedStr := fs.String("rejected", "", "Comma-separated rejected alternatives")
	confidence := fs.Float64("confidence", -1, "Memory Object confidence")

	if err := fs.Parse(args); err != nil {
		fmt.Printf("引数パースエラーですわ: %v\n", err)
		os.Exit(1)
	}

	if *title == "" || *content == "" {
		fmt.Println("エラー: -title と -content は必須指定でしてよ")
		os.Exit(1)
	}

	var tags []string
	if *tagsStr != "" {
		for _, t := range strings.Split(*tagsStr, ",") {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		fmt.Printf("DB接続失敗: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	object := &MemoryObject{Type: *objectType, Scope: *scope, Rationale: *rationale,
		Evidence: splitCommaList(*evidenceStr), RejectedAlternatives: splitCommaList(*rejectedStr)}
	if *confidence >= 0 {
		object.Confidence = confidence
	}
	rec, err := InsertMemoryHeavy(ctx, pool, AddMemoryInput{
		Category:     *cat,
		Title:        *title,
		ContentL0:    *content,
		ContentL1:    *l1,
		ContentL2:    *l2,
		Tags:         tags,
		MemoryObject: object,
	})
	if err != nil {
		fmt.Printf("記憶登録失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✨ 記憶を登録いたしましたわ！\nID: %s\nTitle: %s\nCategory: %s\nClient: %s\n",
		rec.ID, rec.Title, rec.Category, rec.ClientID)
}

// handleSearchHeavy は記憶の横断検索を処理いたしますわ
func handleSearchHeavy(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	query := fs.String("q", "", "Query keyword")
	tag := fs.String("tag", "", "Tag")
	cat := fs.String("cat", "", "Category")
	objectType := fs.String("type", "", "Memory Object type")
	scope := fs.String("scope", "", "Memory Object scope")
	level := fs.Int("level", 1, "Detail level (0:Raw, 1:L1, 2:L2, 3:Tags)")
	limit := fs.Int("limit", 10, "Limit")
	asJSON := fs.Bool("json", false, "Output as JSON")

	if err := fs.Parse(args); err != nil {
		fmt.Printf("引数パースエラーですわ: %v\n", err)
		os.Exit(1)
	}

	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		fmt.Printf("DB接続失敗: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	results, err := SearchMemoriesFilteredHeavy(ctx, pool, *query, *tag, *cat, *objectType, *scope, *limit)
	if err != nil {
		fmt.Printf("検索エラー: %v\n", err)
		os.Exit(1)
	}

	if *asJSON {
		out, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(out))
		return
	}

	if len(results) == 0 {
		fmt.Println("一致する有効な記憶は見つかりませんでしたわ。")
		return
	}

	fmt.Printf("=== 検索結果 (%d 件) ===\n\n", len(results))
	for i, r := range results {
		fmt.Printf("[%d] %s (%s) | ID: %s\n", i+1, r.Title, r.Category, r.ID)
		fmt.Printf("    Tags: %v | Updated: %s | Client: %s\n",
			r.Tags, r.UpdatedAt.Format("2006-01-02 15:04:05"), r.ClientID)

		switch *level {
		case 0:
			fmt.Printf("    [L0 Raw Content]:\n%s\n", r.ContentL0)
		case 1:
			if r.ContentL1 != nil && *r.ContentL1 != "" {
				fmt.Printf("    [L1 Key Points]:\n%s\n", *r.ContentL1)
			} else {
				fmt.Printf("    [L0 Fallback]:\n%s\n", r.ContentL0)
			}
		case 2:
			if r.ContentL2 != nil && *r.ContentL2 != "" {
				fmt.Printf("    [L2 One-liner]: %s\n", *r.ContentL2)
			} else {
				fmt.Printf("    [Title Fallback]: %s\n", r.Title)
			}
		case 3:
			// Tags only
		}
		fmt.Println("------------------------------------------------------------")
	}
}

// handleSupersedeHeavy は自己編集更新処理を遂行いたしますわ
func handleSupersedeHeavy(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("supersede", flag.ExitOnError)
	oldID := fs.String("id", "", "Old Memory UUID to supersede")
	title := fs.String("title", "", "New Title")
	content := fs.String("content", "", "New L0 Content")
	cat := fs.String("cat", "knowledge", "Category")
	l1 := fs.String("l1", "", "L1 Key Points")
	l2 := fs.String("l2", "", "L2 One-liner")
	tagsStr := fs.String("tags", "", "Tags")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *oldID == "" || *title == "" || *content == "" {
		fmt.Println("エラー: -id, -title, -content は必須指定でしてよ")
		os.Exit(1)
	}

	var tags []string
	if *tagsStr != "" {
		for _, t := range strings.Split(*tagsStr, ",") {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		fmt.Printf("DB接続失敗: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	newRec, err := SupersedeMemoryHeavy(ctx, pool, *oldID, AddMemoryInput{
		Category:  *cat,
		Title:     *title,
		ContentL0: *content,
		ContentL1: *l1,
		ContentL2: *l2,
		Tags:      tags,
	})
	if err != nil {
		fmt.Printf("自己編集更新失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔄 記憶を自己編集いたしましたわ！\n旧ID: %s (SUPERSEDED)\n新ID: %s (v%d, ACTIVE)\n",
		*oldID, newRec.ID, newRec.Version)
}

// handleGraphHeavy は知識グラフ関連サブコマンドをルーティングいたしますわ
func handleGraphHeavy(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Println("利用可能なgraphサブコマンド: node, edge, list")
		os.Exit(1)
	}

	sub := args[0]
	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		fmt.Printf("DB接続失敗: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	switch sub {
	case "node":
		fs := flag.NewFlagSet("graph node", flag.ExitOnError)
		name := fs.String("name", "", "Node name")
		nodeType := fs.String("type", "concept", "Node type")
		sum := fs.String("summary", "", "Summary")
		_ = fs.Parse(args[1:])

		if *name == "" {
			fmt.Println("エラー: -name は必須でしてよ")
			os.Exit(1)
		}

		node, err := UpsertNodeHeavy(ctx, pool, *name, *nodeType, *sum, nil)
		if err != nil {
			fmt.Printf("ノード登録エラー: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("🟢 ノード登録完了: %s (%s)\n", node.Name, node.NodeType)

	case "edge":
		fs := flag.NewFlagSet("graph edge", flag.ExitOnError)
		src := fs.String("src", "", "Source node")
		tgt := fs.String("tgt", "", "Target node")
		rel := fs.String("rel", "DEPENDS_ON", "Relation type")
		weight := fs.Float64("weight", 1.0, "Weight")
		evID := fs.String("ev", "", "Evidence memory ID")
		_ = fs.Parse(args[1:])

		if *src == "" || *tgt == "" {
			fmt.Println("エラー: -src と -tgt は必須でしてよ")
			os.Exit(1)
		}

		edge, err := AddEdgeHeavy(ctx, pool, *src, *tgt, *rel, *weight, *evID)
		if err != nil {
			fmt.Printf("エッジ登録エラー: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("🔗 グラフエッジ登録完了: (%s) -[%s]-> (%s)\n", edge.SourceName, edge.RelationType, edge.TargetName)

	case "list":
		edges, err := ListActiveEdgesHeavy(ctx, pool, 50)
		if err != nil {
			fmt.Printf("グラフレコード取得失敗: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("=== 有効な知識グラフ関係性 (%d 件) ===\n", len(edges))
		for _, e := range edges {
			fmt.Printf("・(%s:%s) ──[%s (w=%.1f)]──> (%s:%s)\n",
				e.SourceName, e.SourceType, e.RelationType, e.Weight, e.TargetName, e.TargetType)
		}
	}
}

// handleClientsHeavy は登録端末一覧を表示いたしますわ
func handleClientsHeavy(ctx context.Context) {
	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		fmt.Printf("DB接続失敗: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	rows, err := pool.Query(ctx, "SELECT client_id, hostname, os_info, last_seen_at FROM clients ORDER BY last_seen_at DESC")
	if err != nil {
		fmt.Printf("端末一覧取得エラー: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("=== 登録済みクライアント端末一覧 ===")
	for rows.Next() {
		var cid, host, osInfo string
		var lastSeen time.Time
		_ = rows.Scan(&cid, &host, &osInfo, &lastSeen)
		fmt.Printf("・[%s] Host: %s | OS: %s | LastSeen: %s\n",
			cid, host, osInfo, lastSeen.Format("2006-01-02 15:04:05"))
	}
}

// handleAnalyzeHeavy は diary.md の因果帰属分析と双方向フィードバックを実行いたしますわ
func handleAnalyzeHeavy(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	filePath := fs.String("file", "", "Path to diary.md")
	asJSON := fs.Bool("json", false, "Output in JSON format")
	suggest := fs.Bool("suggest", true, "Show rule optimization diff suggestions")

	if err := fs.Parse(args); err != nil {
		fmt.Printf("引数パースエラーですわ: %v\n", err)
		os.Exit(1)
	}

	target := *filePath
	if target == "" {
		// 公開プロジェクトではカレントディレクトリの diary.md だけを自動検出します。
		candidates := []string{
			"diary.md",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				target = c
				break
			}
		}
	}

	if target == "" {
		fmt.Println("解析対象の diary.md が見つかりませんでしたわ。-file オプションで明示指定してくださいませ。")
		os.Exit(1)
	}

	if err := AnalyzeDiaryFileHeavy(ctx, target, *asJSON, *suggest); err != nil {
		fmt.Printf("解析実行エラーですわ: %v\n", err)
		os.Exit(1)
	}
}
