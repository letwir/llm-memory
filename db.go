// Morphism: RawEnvString → ValidatedDSN ∘ PostgreSQLConnectionPool
// Functor: F(ConfigSpace) ⇒ Category(PostgreSQL_Session)

package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultTimeoutSeconds はDB接続および通常クエリのタイムアウト秒数ですわ
const DefaultTimeoutSeconds = 10

// buildDatabaseURL はビルド時に -ldflags -X で注入できます。
// 環境変数が設定されている場合は環境変数を優先します。
var buildDatabaseURL string

// ParseDSNFromEnvComplex は環境変数からDSN文字列をエレガントに解決いたしますわ
func ParseDSNFromEnvComplex() (string, error) {
	raw := os.Getenv("LLM_MEMORY_DB_URL")
	if raw == "" {
		// プロセス環境変数になければWindowsレジストリ/ユーザー環境変数から探しますの
		raw = os.Getenv("LLM_MEMORY_URL")
	}
	if raw == "" {
		raw = buildDatabaseURL
	}

	if raw == "" {
		return "", fmt.Errorf("DB接続先が未設定ですわ: LLM_MEMORY_DB_URL またはビルド時埋め込み値が必要です")
	}

	if strings.Contains(raw, "://") {
		return raw, nil
	}

	// コロン区切りの接続先形式をパース
	parts := strings.Split(raw, ":")
	if len(parts) < 5 {
		return "", fmt.Errorf("DSNのフォーマットが不正でしてよ: コロン区切りの接続先形式が必要ですわ")
	}

	host := parts[0]
	port := parts[1]
	dbname := parts[2]
	user := parts[3]
	pass := strings.Join(parts[4:], ":") // パスワード中にコロンが含まれる可能性を考慮

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=prefer",
		url.QueryEscape(user),
		url.QueryEscape(pass),
		host,
		port,
		dbname,
	)
	return dsn, nil
}

// GetDBPoolHeavy は接続プールを安全に初期化して返却いたしますわ
func GetDBPoolHeavy(ctx context.Context) (*pgxpool.Pool, error) {
	dsn, err := ParseDSNFromEnvComplex()
	if err != nil {
		return nil, fmt.Errorf("DSN解決エラーですわ: %w", err)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("プール設定のパースに失敗いたしましたわ: %w", err)
	}

	// コネクションプールの上限設定
	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("DBプール作成に失敗いたしましたわ: %w", err)
	}

	// 疎通確認でPingを打ちますわ
	pingCtx, cancel := context.WithTimeout(ctx, DefaultTimeoutSeconds*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("PostgreSQLへのPing疎通に失敗いたしましたわ: %w", err)
	}

	return pool, nil
}
