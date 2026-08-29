package main

import (
	"fmt"
	"strings"
)

// MemoryObject is the durable, normalized meaning of a memory record.
// It lives under metadata so existing databases remain compatible.
type MemoryObject struct {
	Type                 string   `json:"type"`
	Scope                string   `json:"scope"`
	Proposition          string   `json:"proposition,omitempty"`
	Rationale            string   `json:"rationale,omitempty"`
	Evidence             []string `json:"evidence,omitempty"`
	RejectedAlternatives []string `json:"rejected_alternatives,omitempty"`
	Confidence           *float64 `json:"confidence,omitempty"`
	Supersedes           []string `json:"supersedes,omitempty"`
}

var memoryObjectTypes = map[string]bool{
	"invariant": true, "decision": true, "constraint": true, "failure": true,
	"knowledge": true, "procedure": true, "state": true,
}

func defaultMemoryObjectType(category string) string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "decision":
		return "decision"
	case "issue", "state":
		return "state"
	case "diary", "failure":
		return "failure"
	case "method", "procedure":
		return "procedure"
	case "invariant":
		return "invariant"
	case "constraint":
		return "constraint"
	default:
		return "knowledge"
	}
}

func normalizeMemoryObject(category, title, content string, object *MemoryObject) (MemoryObject, error) {
	result := MemoryObject{Type: defaultMemoryObjectType(category), Scope: "project", Proposition: strings.TrimSpace(content)}
	if object != nil {
		result = *object
		result.Type = strings.ToLower(strings.TrimSpace(result.Type))
		result.Scope = strings.ToLower(strings.TrimSpace(result.Scope))
		result.Proposition = strings.TrimSpace(result.Proposition)
		result.Rationale = strings.TrimSpace(result.Rationale)
	}
	if result.Type == "" {
		result.Type = defaultMemoryObjectType(category)
	}
	if !memoryObjectTypes[result.Type] {
		return MemoryObject{}, fmt.Errorf("memory_object.type %q is invalid", result.Type)
	}
	if result.Scope == "" {
		result.Scope = "project"
	}
	if result.Scope != "global" && result.Scope != "project" && !strings.HasPrefix(result.Scope, "subsystem:") {
		return MemoryObject{}, fmt.Errorf("memory_object.scope %q is invalid", result.Scope)
	}
	if result.Proposition == "" {
		result.Proposition = strings.TrimSpace(title)
	}
	if result.Confidence != nil && (*result.Confidence < 0 || *result.Confidence > 1) {
		return MemoryObject{}, fmt.Errorf("memory_object.confidence must be between 0 and 1")
	}
	return result, nil
}

func normalizeMemoryMetadata(metadata map[string]interface{}, category, title, content string, object *MemoryObject) (map[string]interface{}, MemoryObject, error) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	result, err := normalizeMemoryObject(category, title, content, object)
	if err != nil {
		return nil, MemoryObject{}, err
	}
	metadata["memory_object"] = result
	return metadata, result, nil
}
