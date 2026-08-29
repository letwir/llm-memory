// Morphism: HostOSState → ClientRecord ∘ RegisteredDBState
// Functor: F(LocalHost) ⇒ Category(ClientsTable)

package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ClientInfo は端末識別構造体ですわ
type ClientInfo struct {
	ClientID    string    `json:"client_id"`
	Hostname    string    `json:"hostname"`
	OSInfo      string    `json:"os_info"`
	Description string    `json:"description"`
	LastSeenAt  time.Time `json:"last_seen_at"`
}

// ResolveLocalClientID は環境変数またはホスト名からクライアントIDを決定いたしますわ
func ResolveLocalClientID() string {
	cid := os.Getenv("LLM_MEMORY_CLIENT_ID")
	if cid != "" {
		return cid
	}

	host, err := os.Hostname()
	if err == nil && host != "" {
		return fmt.Sprintf("%s-%s", runtime.GOOS, host)
	}

	return fmt.Sprintf("node-%s", runtime.GOOS)
}

// AutoRegisterClientHeavy は起動時に端末情報を自動でアップサートいたしますの
func AutoRegisterClientHeavy(ctx context.Context, pool *pgxpool.Pool) (*ClientInfo, error) {
	clientID := ResolveLocalClientID()
	hostname, _ := os.Hostname()
	osInfo := fmt.Sprintf("%s_%s_%s", runtime.GOOS, runtime.GOARCH, runtime.Version())

	query := `
		INSERT INTO clients (client_id, hostname, os_info, description, last_seen_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (client_id) DO UPDATE 
		SET hostname = EXCLUDED.hostname,
		    os_info = EXCLUDED.os_info,
		    last_seen_at = NOW()
		RETURNING client_id, hostname, os_info, description, last_seen_at;
	`

	var info ClientInfo
	var desc *string

	err := pool.QueryRow(ctx, query, clientID, hostname, osInfo, "Auto-registered from CLI").Scan(
		&info.ClientID,
		&info.Hostname,
		&info.OSInfo,
		&desc,
		&info.LastSeenAt,
	)
	if err != nil {
		return nil, fmt.Errorf("クライアント情報の登録に失敗いたしましたわ: %w", err)
	}

	if desc != nil {
		info.Description = *desc
	}

	return &info, nil
}
