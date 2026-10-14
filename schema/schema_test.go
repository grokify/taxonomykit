package schema

import (
	"encoding/json"
	"testing"
)

func TestEmbeddedSchema(t *testing.T) {
	var s struct {
		Schema string                     `json:"$schema"`
		ID     string                     `json:"$id"`
		Ref    string                     `json:"$ref"`
		Defs   map[string]json.RawMessage `json:"$defs"`
	}
	if err := json.Unmarshal(Document(), &s); err != nil {
		t.Fatal(err)
	}
	if s.ID != ID {
		t.Fatalf("$id %q", s.ID)
	}
	for _, def := range []string{"Document", "Node", "Label"} {
		if _, ok := s.Defs[def]; !ok {
			t.Fatalf("missing $defs.%s", def)
		}
	}
	var node struct {
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if err := json.Unmarshal(s.Defs["Node"], &node); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"id", "parents", "labels", "order", "attrs"} {
		if _, ok := node.Properties[p]; !ok {
			t.Fatalf("Node is missing property %q", p)
		}
	}
	if len(node.Required) != 2 || node.Required[0] != "id" || node.Required[1] != "labels" {
		t.Fatalf("Node required: %v", node.Required)
	}
	// Document() must return a copy.
	Document()[0] = 'x'
	if Document()[0] != '{' {
		t.Fatal("Document must not expose the embedded bytes")
	}
}
