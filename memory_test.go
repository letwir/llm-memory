package main

import (
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestBuildSearchMemoriesQueryFallsBackWithoutSearchDocument(t *testing.T) {
	withFTS, withArgs := buildSearchMemoriesQuery("database error", "", "", "", "", 10, true)
	withoutFTS, withoutArgs := buildSearchMemoriesQuery("database error", "", "", "", "", 10, false)

	if !strings.Contains(withFTS, "search_document @@") {
		t.Fatalf("full-text query does not include search_document: %s", withFTS)
	}
	if strings.Contains(withoutFTS, "search_document") {
		t.Fatalf("fallback query still includes search_document: %s", withoutFTS)
	}
	if len(withArgs) != 2 || len(withoutArgs) != 2 {
		t.Fatalf("unexpected argument counts: full-text=%d fallback=%d", len(withArgs), len(withoutArgs))
	}
}

func TestIsMissingSearchDocumentError(t *testing.T) {
	if !isMissingSearchDocumentError(&pgconn.PgError{Code: "42703", Message: "column search_document does not exist"}) {
		t.Fatal("search_document missing-column error was not recognized")
	}
	if isMissingSearchDocumentError(&pgconn.PgError{Code: "42703", Message: "column content_l0 does not exist"}) {
		t.Fatal("unrelated missing-column error was incorrectly recognized")
	}
}
