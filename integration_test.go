//go:build integration

package main

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

func TestConcurrentActiveIdentityInsert(t *testing.T) {
	if os.Getenv("LLM_MEMORY_INTEGRATION") != "1" {
		t.Skip("set LLM_MEMORY_INTEGRATION=1 to run against the configured PostgreSQL database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	title := "integration-concurrency-" + time.Now().UTC().Format("20060102T150405.000000000")
	const attempts = 8
	results := make(chan error, attempts)
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := InsertMemoryHeavy(ctx, pool, AddMemoryInput{
				Category:  "integration",
				Title:     title,
				ContentL0: "concurrent identity test",
			})
			results <- err
		}(i)
	}
	wg.Wait()
	close(results)

	var successes int
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent inserts succeeded %d times; want exactly 1", successes)
	}

	_, err = pool.Exec(ctx, "UPDATE memories SET status = 'DEPRECATED', valid_to = NOW(), tx_invalidated_at = NOW() WHERE category = $1 AND title = $2", "integration", title)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWORMBoundaryTrigger(t *testing.T) {
	if os.Getenv("LLM_MEMORY_INTEGRATION") != "1" {
		t.Skip("set LLM_MEMORY_INTEGRATION=1 to run against the configured PostgreSQL database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := GetDBPoolHeavy(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	var triggerExists bool
	err = pool.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM pg_trigger
		WHERE tgrelid = 'memories'::regclass
		  AND tgname = 'memories_immutable_fields_trigger'
	)`).Scan(&triggerExists)
	if err != nil {
		t.Fatal(err)
	}
	if !triggerExists {
		t.Skip("010_worm_boundary.sql has not been applied")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	var clientID string
	if err := tx.QueryRow(ctx, "SELECT client_id FROM clients LIMIT 1").Scan(&clientID); err != nil {
		t.Fatal(err)
	}
	var memoryID string
	err = tx.QueryRow(ctx, `INSERT INTO memories(client_id, category, title, content_l0)
		VALUES ($1, 'integration', 'worm-boundary-test', 'immutable raw') RETURNING id`, clientID).Scan(&memoryID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "SAVEPOINT immutable_update"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "UPDATE memories SET title = 'changed' WHERE id = $1", memoryID); err == nil {
		t.Fatal("immutable title update unexpectedly succeeded")
	}
	if _, err := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT immutable_update"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "UPDATE memories SET status = 'DEPRECATED', valid_to = NOW(), tx_invalidated_at = NOW() WHERE id = $1", memoryID); err != nil {
		t.Fatalf("lifecycle update rejected: %v", err)
	}
	if _, err := tx.Exec(ctx, "SAVEPOINT delete_attempt"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM memories WHERE id = $1", memoryID); err == nil {
		t.Fatal("DELETE unexpectedly succeeded")
	}
	if _, err := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT delete_attempt"); err != nil {
		t.Fatal(err)
	}
}
