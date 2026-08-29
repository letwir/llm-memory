package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultEmbeddingDimensions = 768

func embeddingModel() string {
	if model := os.Getenv("LLM_MEMORY_EMBEDDING_MODEL"); model != "" {
		return model
	}
	return "gemini-embedding-001"
}

type embeddingResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

// EmbedTextHeavy uses Gemini's retrieval embedding endpoint. It is optional and never part of normal ingest.
func EmbedTextHeavy(ctx context.Context, text, taskType string) ([]float32, error) {
	key := resolveGeminiAPIKey()
	if key == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required for semantic embeddings")
	}
	model := embeddingModel()
	payload := map[string]interface{}{
		"content":              map[string]interface{}{"parts": []map[string]string{{"text": text}}},
		"taskType":             taskType,
		"outputDimensionality": defaultEmbeddingDimensions,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://generativelanguage.googleapis.com/v1beta/models/"+model+":embedContent",
		strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", key)
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding API: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("embedding response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	var result embeddingResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("embedding JSON: %w", err)
	}
	if len(result.Embedding.Values) != defaultEmbeddingDimensions {
		return nil, fmt.Errorf("embedding dimensions=%d, want %d", len(result.Embedding.Values), defaultEmbeddingDimensions)
	}
	return result.Embedding.Values, nil
}

func UpsertMemoryEmbeddingHeavy(ctx context.Context, pool *pgxpool.Pool, memoryID, title, content string) error {
	vector, err := EmbedTextHeavy(ctx, "title: "+title+" | text: "+content, "RETRIEVAL_DOCUMENT")
	if err != nil {
		return err
	}
	buf := make([]string, len(vector))
	for i, value := range vector {
		buf[i] = strconv.FormatFloat(float64(value), 'g', -1, 32)
	}
	hash := sha256.Sum256([]byte(content))
	_, err = pool.Exec(ctx, `
		INSERT INTO memory_embeddings(memory_id, model, dimensions, embedding, source_sha256)
		VALUES ($1, $2, $3, $4::vector, $5)
		ON CONFLICT (memory_id, model) DO UPDATE SET
		  dimensions = EXCLUDED.dimensions,
		  embedding = EXCLUDED.embedding,
		  source_sha256 = EXCLUDED.source_sha256,
		  created_at = NOW()`,
		memoryID, embeddingModel(), defaultEmbeddingDimensions,
		"["+strings.Join(buf, ",")+"]", hex.EncodeToString(hash[:]))
	return err
}

func SearchSemanticMemoriesHeavy(ctx context.Context, pool *pgxpool.Pool, query string, limit int) ([]MemoryRecord, error) {
	vector, err := EmbedTextHeavy(ctx, query, "RETRIEVAL_QUERY")
	if err != nil {
		return nil, err
	}
	buf := make([]string, len(vector))
	for i, value := range vector {
		buf[i] = strconv.FormatFloat(float64(value), 'g', -1, 32)
	}
	if limit <= 0 {
		limit = 10
	}
	rows, err := pool.Query(ctx, `
		SELECT m.id, m.client_id, m.category, m.title, m.content_l0, m.content_l1, m.content_l2,
		       m.tags, m.current_level, m.valid_from, m.valid_to, m.version, m.metadata,
		       m.created_at, m.updated_at
		FROM v_active_memories m
		JOIN memory_embeddings e ON e.memory_id = m.id
		WHERE e.model = $2
		ORDER BY e.embedding <=> $1::vector
		LIMIT $3`, "["+strings.Join(buf, ",")+"]", embeddingModel(), limit)
	if err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}
	defer rows.Close()
	var records []MemoryRecord
	for rows.Next() {
		var rec MemoryRecord
		var rawMeta []byte
		if err := rows.Scan(&rec.ID, &rec.ClientID, &rec.Category, &rec.Title, &rec.ContentL0, &rec.ContentL1, &rec.ContentL2, &rec.Tags, &rec.CurrentLevel, &rec.ValidFrom, &rec.ValidTo, &rec.Version, &rawMeta, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(rawMeta, &rec.Metadata)
		rec.Status = "ACTIVE"
		records = append(records, rec)
	}
	return records, rows.Err()
}
