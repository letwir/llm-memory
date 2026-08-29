package main

import "testing"

func TestNormalizeMemoryObjectDefaultsFromCategory(t *testing.T) {
	metadata, object, err := normalizeMemoryMetadata(nil, "decision", "Choose JSONB", "Store full precision in JSONB", nil)
	if err != nil {
		t.Fatal(err)
	}
	if object.Type != "decision" || object.Scope != "project" {
		t.Fatalf("unexpected object: %+v", object)
	}
	if metadata["memory_object"] == nil {
		t.Fatal("normalized object was not stored in metadata")
	}
}

func TestNormalizeMemoryObjectRejectsInvalidValues(t *testing.T) {
	confidence := 1.1
	_, _, err := normalizeMemoryMetadata(nil, "knowledge", "title", "content", &MemoryObject{Type: "unknown"})
	if err == nil {
		t.Fatal("invalid type must fail")
	}
	_, _, err = normalizeMemoryMetadata(nil, "knowledge", "title", "content", &MemoryObject{Confidence: &confidence})
	if err == nil {
		t.Fatal("invalid confidence must fail")
	}
}
