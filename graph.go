// Morphism: EntityRelations → KnowledgeGraphStructure ∘ MutatedPostgresGraph
// Functor: F(SemanticGraph) ⇒ Category(KnowledgeEdges_Nodes)

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NodeRecord はグラフノード構造体ですわ
type NodeRecord struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	NodeType   string                 `json:"node_type"`
	Summary    string                 `json:"summary"`
	Attributes map[string]interface{} `json:"attributes"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// EdgeRecord はグラフトリプル構造体ですわ
type EdgeRecord struct {
	EdgeID           string     `json:"edge_id"`
	SourceName       string     `json:"source_name"`
	SourceType       string     `json:"source_type"`
	RelationType     string     `json:"relation_type"`
	TargetName       string     `json:"target_name"`
	TargetType       string     `json:"target_type"`
	Weight           float64    `json:"weight"`
	ValidFrom        time.Time  `json:"valid_from"`
	ValidTo          *time.Time `json:"valid_to,omitempty"`
	EvidenceMemoryID *string    `json:"evidence_memory_id,omitempty"`
}

// UpsertNodeHeavy はエンティティノードを登録または更新いたしますわ
func UpsertNodeHeavy(ctx context.Context, pool *pgxpool.Pool, name string, nodeType string, summary string, attrs map[string]interface{}) (*NodeRecord, error) {
	if name == "" {
		return nil, fmt.Errorf("ノード名は必須でしてよ")
	}
	if nodeType == "" {
		nodeType = "concept"
	}
	if attrs == nil {
		attrs = make(map[string]interface{})
	}

	attrsJSON, _ := json.Marshal(attrs)

	query := `
		INSERT INTO knowledge_nodes (name, node_type, summary, attributes, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (name) DO UPDATE 
		SET node_type = EXCLUDED.node_type,
		    summary = COALESCE(NULLIF(EXCLUDED.summary, ''), knowledge_nodes.summary),
		    attributes = knowledge_nodes.attributes || EXCLUDED.attributes,
		    updated_at = NOW()
		RETURNING id, name, node_type, summary, attributes, created_at, updated_at;
	`

	var rec NodeRecord
	var rawAttrs []byte
	var sum *string

	err := pool.QueryRow(ctx, query, name, nodeType, summary, attrsJSON).Scan(
		&rec.ID, &rec.Name, &rec.NodeType, &sum, &rawAttrs, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ノードのアップサートに失敗いたしましたわ: %w", err)
	}

	if sum != nil {
		rec.Summary = *sum
	}
	_ = json.Unmarshal(rawAttrs, &rec.Attributes)
	return &rec, nil
}

// AddEdgeHeavy は2つのノード間に時間軸関係性エッジを結びますわ
func AddEdgeHeavy(ctx context.Context, pool *pgxpool.Pool, sourceName string, targetName string, relationType string, weight float64, evidenceID string) (*EdgeRecord, error) {
	if sourceName == "" || targetName == "" || relationType == "" {
		return nil, fmt.Errorf("始点ノード、終点ノード、関係性タイプはすべて必須でしてよ")
	}
	if weight <= 0 {
		weight = 1.0
	}

	// 始点・終点ノードを自動確保
	srcNode, err := UpsertNodeHeavy(ctx, pool, sourceName, "concept", "", nil)
	if err != nil {
		return nil, fmt.Errorf("始点ノード確保失敗ですわ: %w", err)
	}

	tgtNode, err := UpsertNodeHeavy(ctx, pool, targetName, "concept", "", nil)
	if err != nil {
		return nil, fmt.Errorf("終点ノード確保失敗ですわ: %w", err)
	}

	var evID *string
	if evidenceID != "" {
		evID = &evidenceID
	}

	query := `
		INSERT INTO knowledge_edges (
			source_node_id, target_node_id, relation_type, weight, evidence_memory_id
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, valid_from, valid_to;
	`

	var edgeID string
	var validFrom time.Time
	var validTo *time.Time

	err = pool.QueryRow(ctx, query, srcNode.ID, tgtNode.ID, relationType, weight, evID).Scan(
		&edgeID, &validFrom, &validTo,
	)
	if err != nil {
		return nil, fmt.Errorf("エッジ挿入に失敗いたしましたわ: %w", err)
	}

	return &EdgeRecord{
		EdgeID:           edgeID,
		SourceName:       srcNode.Name,
		SourceType:       srcNode.NodeType,
		RelationType:     relationType,
		TargetName:       tgtNode.Name,
		TargetType:       tgtNode.NodeType,
		Weight:           weight,
		ValidFrom:        validFrom,
		ValidTo:          validTo,
		EvidenceMemoryID: evID,
	}, nil
}

// ListActiveEdgesHeavy は有効な知識グラフの関係性一覧を取得いたしますわ
func ListActiveEdgesHeavy(ctx context.Context, pool *pgxpool.Pool, limit int) ([]EdgeRecord, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT 
			edge_id, source_name, source_type, relation_type, 
			target_name, target_type, weight, valid_from, valid_to, evidence_memory_id
		FROM v_active_knowledge_graph
		ORDER BY valid_from DESC
		LIMIT $1;
	`

	rows, err := pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("グラフ取得クエリ失敗ですわ: %w", err)
	}
	defer rows.Close()

	var edges []EdgeRecord
	for rows.Next() {
		var e EdgeRecord
		err := rows.Scan(
			&e.EdgeID, &e.SourceName, &e.SourceType, &e.RelationType,
			&e.TargetName, &e.TargetType, &e.Weight, &e.ValidFrom, &e.ValidTo, &e.EvidenceMemoryID,
		)
		if err != nil {
			return nil, fmt.Errorf("グラフレコードスキャン失敗ですわ: %w", err)
		}
		edges = append(edges, e)
	}

	return edges, nil
}
