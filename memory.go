// Morphism: InputMemory ∘ BiTemporalContext → StoredMemory ∘ MutatedPostgresState
// Functor: F(CognitiveBuffer) ⇒ Category(BiTemporalMemories)

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MemoryRecord は記憶の完全な構造体ですわ
type MemoryRecord struct {
	ID              string                 `json:"id"`
	ClientID        string                 `json:"client_id"`
	Category        string                 `json:"category"`
	Title           string                 `json:"title"`
	ContentL0       string                 `json:"content_l0"`
	ContentL1       *string                `json:"content_l1,omitempty"`
	ContentL2       *string                `json:"content_l2,omitempty"`
	Tags            []string               `json:"tags"`
	CurrentLevel    int                    `json:"current_level"`
	ValidFrom       time.Time              `json:"valid_from"`
	ValidTo         *time.Time             `json:"valid_to,omitempty"`
	TxCreatedAt     time.Time              `json:"tx_created_at"`
	TxInvalidatedAt *time.Time             `json:"tx_invalidated_at,omitempty"`
	Status          string                 `json:"status"`
	SupersededBy    *string                `json:"superseded_by,omitempty"`
	Version         int                    `json:"version"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// AddMemoryInput は新規記憶登録時のパラメータですわ
type AddMemoryInput struct {
	Category     string
	Title        string
	ContentL0    string
	ContentL1    string
	ContentL2    string
	Tags         []string
	Metadata     map[string]interface{}
	MemoryObject *MemoryObject
}

// normalizeIdentity removes case and insignificant whitespace for ingest identity checks.
func normalizeIdentity(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

// InsertMemoryHeavy は新規記憶を二重時間軸テーブルへ登録いたしますわ
func InsertMemoryHeavy(ctx context.Context, pool *pgxpool.Pool, in AddMemoryInput) (*MemoryRecord, error) {
	var err error
	if strings.TrimSpace(in.Title) == "" {
		return nil, fmt.Errorf("タイトルの指定は必須でしてよ")
	}
	if strings.TrimSpace(in.ContentL0) == "" {
		return nil, fmt.Errorf("本文 (L0) の指定は必須でしてよ")
	}
	if in.Category == "" {
		in.Category = "knowledge"
	}
	in.Metadata, _, err = normalizeMemoryMetadata(in.Metadata, in.Category, in.Title, in.ContentL0, in.MemoryObject)
	if err != nil {
		return nil, fmt.Errorf("memory object validation failed: %w", err)
	}

	clientID := ResolveLocalClientID()
	metaJSON, err := json.Marshal(in.Metadata)
	if err != nil {
		return nil, fmt.Errorf("メタデータのJSON変換に失敗いたしましたわ: %w", err)
	}

	query := `
		INSERT INTO memories (
			client_id, category, title, content_l0, content_l1, content_l2, tags, metadata
		) VALUES (
			$1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, $8
		) RETURNING 
			id, client_id, category, title, content_l0, content_l1, content_l2, tags, 
			current_level, valid_from, valid_to, tx_created_at, tx_invalidated_at, 
			status, superseded_by, version, metadata, created_at, updated_at;
	`

	var rec MemoryRecord
	var rawMeta []byte

	err = pool.QueryRow(ctx, query,
		clientID, in.Category, in.Title, in.ContentL0, in.ContentL1, in.ContentL2, in.Tags, metaJSON,
	).Scan(
		&rec.ID, &rec.ClientID, &rec.Category, &rec.Title,
		&rec.ContentL0, &rec.ContentL1, &rec.ContentL2, &rec.Tags,
		&rec.CurrentLevel, &rec.ValidFrom, &rec.ValidTo, &rec.TxCreatedAt, &rec.TxInvalidatedAt,
		&rec.Status, &rec.SupersededBy, &rec.Version, &rawMeta,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("記憶の挿入クエリ実行に失敗いたしましたわ: %w", err)
	}

	_ = json.Unmarshal(rawMeta, &rec.Metadata)
	return &rec, nil
}

// SearchMemoriesHeavy はキーワード・タグ・カテゴリを組み合わせて高速横断検索いたしますわ
func SearchMemoriesHeavy(ctx context.Context, pool *pgxpool.Pool, keyword string, tag string, category string, limit int) ([]MemoryRecord, error) {
	return SearchMemoriesFilteredHeavy(ctx, pool, keyword, tag, category, "", "", limit)
}

// SearchMemoriesFilteredHeavy adds normalized Memory Object filters while keeping the original API intact.
func SearchMemoriesFilteredHeavy(ctx context.Context, pool *pgxpool.Pool, keyword string, tag string, category string, objectType string, scope string, limit int) ([]MemoryRecord, error) {
	query, args := buildSearchMemoriesQuery(keyword, tag, category, objectType, scope, limit, true)
	rows, err := pool.Query(ctx, query, args...)
	if err != nil && keyword != "" && isMissingSearchDocumentError(err) {
		// Older deployed databases predate the optional full-text column. Preserve
		// keyword search rather than requiring a schema mutation for read-only use.
		query, args = buildSearchMemoriesQuery(keyword, tag, category, objectType, scope, limit, false)
		rows, err = pool.Query(ctx, query, args...)
	}
	if err != nil {
		return nil, fmt.Errorf("検索クエリ実行エラーでしてよ: %w", err)
	}
	defer rows.Close()

	var records []MemoryRecord
	for rows.Next() {
		var rec MemoryRecord
		var rawMeta []byte
		err := rows.Scan(
			&rec.ID, &rec.ClientID, &rec.Category, &rec.Title,
			&rec.ContentL0, &rec.ContentL1, &rec.ContentL2, &rec.Tags,
			&rec.CurrentLevel, &rec.ValidFrom, &rec.ValidTo, &rec.Version,
			&rawMeta, &rec.CreatedAt, &rec.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("レコードのスキャンに失敗いたしましたわ: %w", err)
		}
		_ = json.Unmarshal(rawMeta, &rec.Metadata)
		rec.Status = "ACTIVE"
		records = append(records, rec)
	}

	return records, nil
}

func buildSearchMemoriesQuery(keyword string, tag string, category string, objectType string, scope string, limit int, useSearchDocument bool) (string, []interface{}) {
	if limit <= 0 {
		limit = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	// 有効ビュー (v_active_memories) から検索
	baseQuery := `
		SELECT 
			id, client_id, category, title, content_l0, content_l1, content_l2, tags, 
			current_level, valid_from, valid_to, version, metadata, created_at, updated_at
		FROM v_active_memories
	`

	if category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}
	if objectType != "" {
		conditions = append(conditions, fmt.Sprintf("COALESCE(metadata->'memory_object'->>'type', CASE category WHEN 'decision' THEN 'decision' WHEN 'issue' THEN 'state' WHEN 'diary' THEN 'failure' WHEN 'method' THEN 'procedure' WHEN 'invariant' THEN 'invariant' WHEN 'constraint' THEN 'constraint' ELSE 'knowledge' END) = $%d", argIdx))
		args = append(args, objectType)
		argIdx++
	}
	if scope != "" {
		conditions = append(conditions, fmt.Sprintf("COALESCE(metadata->'memory_object'->>'scope', 'project') = $%d", argIdx))
		args = append(args, scope)
		argIdx++
	}

	if tag != "" {
		conditions = append(conditions, fmt.Sprintf("tags @> ARRAY[$%d]", argIdx))
		args = append(args, tag)
		argIdx++
	}

	if keyword != "" {
		// Full-text search handles token matches; ILIKE preserves substring and Japanese fallback behavior.
		if useSearchDocument {
			conditions = append(conditions, fmt.Sprintf("(search_document @@ websearch_to_tsquery('simple', $%d) OR title ILIKE '%%' || $%d || '%%' OR content_l0 ILIKE '%%' || $%d || '%%')", argIdx, argIdx, argIdx))
		} else {
			conditions = append(conditions, fmt.Sprintf("(title ILIKE '%%' || $%d || '%%' OR content_l0 ILIKE '%%' || $%d || '%%')", argIdx, argIdx))
		}
		args = append(args, keyword)
		argIdx++
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	baseQuery += fmt.Sprintf(" ORDER BY updated_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	return baseQuery, args
}

func isMissingSearchDocumentError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42703" && strings.Contains(pgErr.Message, "search_document")
}

// SupersedeMemoryHeavy は古い記憶を無効化し、後続の新しい記憶をトランザクション内で安全に作成いたしますわ
func SupersedeMemoryHeavy(ctx context.Context, pool *pgxpool.Pool, oldID string, in AddMemoryInput) (*MemoryRecord, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始失敗ですわ: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. 旧レコードのバージョン取得と存在確認
	var oldVer int
	var oldStatus string
	err = tx.QueryRow(ctx, "SELECT version, status FROM memories WHERE id = $1 FOR UPDATE", oldID).Scan(&oldVer, &oldStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("指定されたID (%s) の記憶が存在いたしませんわ", oldID)
		}
		return nil, fmt.Errorf("旧記憶のロック取得に失敗いたしましたわ: %w", err)
	}
	if oldStatus != "ACTIVE" {
		return nil, fmt.Errorf("旧記憶の状態がACTIVEではありません (id=%s, status=%s)", oldID, oldStatus)
	}

	clientID := ResolveLocalClientID()
	in.Metadata, _, err = normalizeMemoryMetadata(in.Metadata, in.Category, in.Title, in.ContentL0, in.MemoryObject)
	if err != nil {
		return nil, fmt.Errorf("memory object validation failed: %w", err)
	}
	metaJSON, err := json.Marshal(in.Metadata)
	if err != nil {
		return nil, fmt.Errorf("メタデータのJSON変換に失敗いたしましたわ: %w", err)
	}

	// 2. 新レコードの作成 (version = oldVer + 1)
	insertQuery := `
		INSERT INTO memories (
			client_id, category, title, content_l0, content_l1, content_l2, tags, version, metadata
		) VALUES (
			$1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, $8, $9
		) RETURNING 
			id, client_id, category, title, content_l0, content_l1, content_l2, tags, 
			current_level, valid_from, valid_to, tx_created_at, tx_invalidated_at, 
			status, superseded_by, version, metadata, created_at, updated_at;
	`

	var newRec MemoryRecord
	var rawMeta []byte

	err = tx.QueryRow(ctx, insertQuery,
		clientID, in.Category, in.Title, in.ContentL0, in.ContentL1, in.ContentL2, in.Tags, oldVer+1, metaJSON,
	).Scan(
		&newRec.ID, &newRec.ClientID, &newRec.Category, &newRec.Title,
		&newRec.ContentL0, &newRec.ContentL1, &newRec.ContentL2, &newRec.Tags,
		&newRec.CurrentLevel, &newRec.ValidFrom, &newRec.ValidTo, &newRec.TxCreatedAt, &newRec.TxInvalidatedAt,
		&newRec.Status, &newRec.SupersededBy, &newRec.Version, &rawMeta,
		&newRec.CreatedAt, &newRec.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("新記憶の作成に失敗いたしましたわ: %w", err)
	}

	// 3. 旧レコードを SUPERSEDED & valid_to 更新
	updateQuery := `
		UPDATE memories 
		SET status = 'SUPERSEDED',
		    valid_to = NOW(),
		    tx_invalidated_at = NOW(),
		    superseded_by = $1,
		    updated_at = NOW()
		WHERE id = $2;
	`
	_, err = tx.Exec(ctx, updateQuery, newRec.ID, oldID)
	if err != nil {
		return nil, fmt.Errorf("旧記憶の無効化更新に失敗いたしましたわ: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("トランザクションコミットに失敗いたしましたわ: %w", err)
	}

	_ = json.Unmarshal(rawMeta, &newRec.Metadata)
	return &newRec, nil
}
