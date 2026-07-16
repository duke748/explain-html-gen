package mermaid

import (
	"strings"
	"testing"
)

func TestParseValidFlowchart(t *testing.T) {
	source := `flowchart LR
Reader[Reader] -->|opens| HTML[Generated HTML]
HTML --> Quiz[Quiz tab]`

	diagram, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if diagram.Direction != "LR" {
		t.Fatalf("Direction = %q, want LR", diagram.Direction)
	}
	if len(diagram.Nodes) != 3 {
		t.Fatalf("Nodes = %d, want 3", len(diagram.Nodes))
	}
	if len(diagram.Edges) != 2 {
		t.Fatalf("Edges = %d, want 2", len(diagram.Edges))
	}
	if diagram.Edges[0].Label != "opens" {
		t.Fatalf("First edge label = %q, want opens", diagram.Edges[0].Label)
	}
}

func TestParseRejectsUnsupportedSyntax(t *testing.T) {
	source := `graph TD
subgraph API
A --> B
end`

	err := Validate(source)
	if err == nil {
		t.Fatal("Validate returned nil, want unsupported syntax error")
	}
	if !strings.Contains(err.Error(), "unsupported mermaid syntax") {
		t.Fatalf("Error = %q, want unsupported syntax message", err.Error())
	}
}
