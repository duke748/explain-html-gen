package mermaid

import (
	"fmt"
	"regexp"
	"strings"
)

// Diagram is the parsed subset of Mermaid syntax this generator can render
// without external JavaScript dependencies.
type Diagram struct {
	Direction string
	Nodes     []Node
	Edges     []Edge
}

type Node struct {
	ID    string
	Label string
}

type Edge struct {
	From  string
	To    string
	Label string
}

var (
	headerPattern = regexp.MustCompile(`^(?:graph|flowchart)\s+(TD|TB|BT|LR|RL)\s*;?$`)
	edgePattern   = regexp.MustCompile(`^(.+?)\s*(?:-->|---|==>)\s*(?:\|([^|]+)\|\s*)?(.+?)\s*;?$`)
	nodePattern   = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9_-]*)(?:\[(.+)\]|\((.+)\)|\{(.+)\})?$`)
	idPattern     = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)
)

// Validate returns a clear error when source uses Mermaid syntax this renderer
// cannot prove it understands.
func Validate(source string) error {
	_, err := Parse(source)
	return err
}

// Parse accepts a practical subset of Mermaid flowchart syntax:
//
//	graph TD
//	A[Reader] -->|opens| B[HTML]
//	B --> C[Quiz]
func Parse(source string) (*Diagram, error) {
	lines := meaningfulLines(source)
	if len(lines) == 0 {
		return nil, fmt.Errorf("mermaid diagram is empty")
	}

	header := headerPattern.FindStringSubmatch(lines[0])
	if header == nil {
		return nil, fmt.Errorf("mermaid diagram must start with 'graph <direction>' or 'flowchart <direction>'")
	}

	diagram := &Diagram{
		Direction: header[1],
		Nodes:     make([]Node, 0),
		Edges:     make([]Edge, 0),
	}
	nodeLabels := make(map[string]string)
	nodeOrder := make([]string, 0)

	ensureNode := func(raw string) (string, error) {
		id, label, err := parseNode(raw)
		if err != nil {
			return "", err
		}
		if _, exists := nodeLabels[id]; !exists {
			nodeOrder = append(nodeOrder, id)
			nodeLabels[id] = label
		} else if label != id {
			nodeLabels[id] = label
		}
		return id, nil
	}

	for i, line := range lines[1:] {
		lineNo := i + 2
		if unsupportedMermaidLine(line) {
			return nil, fmt.Errorf("unsupported mermaid syntax on line %d: %s", lineNo, line)
		}

		if match := edgePattern.FindStringSubmatch(line); match != nil {
			from, err := ensureNode(match[1])
			if err != nil {
				return nil, fmt.Errorf("invalid mermaid source node on line %d: %w", lineNo, err)
			}
			to, err := ensureNode(match[3])
			if err != nil {
				return nil, fmt.Errorf("invalid mermaid target node on line %d: %w", lineNo, err)
			}
			diagram.Edges = append(diagram.Edges, Edge{
				From:  from,
				To:    to,
				Label: cleanLabel(match[2]),
			})
			continue
		}

		id, label, err := parseNode(line)
		if err != nil {
			return nil, fmt.Errorf("unsupported mermaid syntax on line %d: %s", lineNo, line)
		}
		if _, exists := nodeLabels[id]; !exists {
			nodeOrder = append(nodeOrder, id)
		}
		nodeLabels[id] = label
	}

	if len(diagram.Edges) == 0 {
		return nil, fmt.Errorf("mermaid diagram must include at least one edge")
	}

	for _, id := range nodeOrder {
		diagram.Nodes = append(diagram.Nodes, Node{ID: id, Label: nodeLabels[id]})
	}
	return diagram, nil
}

func meaningfulLines(source string) []string {
	var lines []string
	for _, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		lines = append(lines, trimmed)
	}
	return lines
}

func parseNode(raw string) (string, string, error) {
	raw = strings.TrimSpace(strings.TrimSuffix(raw, ";"))
	match := nodePattern.FindStringSubmatch(raw)
	if match == nil {
		if idPattern.MatchString(raw) {
			return raw, raw, nil
		}
		return "", "", fmt.Errorf("expected node ID or node label expression, got %q", raw)
	}

	id := match[1]
	label := id
	for _, candidate := range match[2:] {
		if candidate != "" {
			label = cleanLabel(candidate)
			break
		}
	}
	return id, label, nil
}

func cleanLabel(label string) string {
	label = strings.TrimSpace(label)
	label = strings.Trim(label, `"'`)
	return label
}

func unsupportedMermaidLine(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	return strings.HasPrefix(lower, "subgraph ") ||
		lower == "end" ||
		strings.Contains(lower, ":::") ||
		strings.Contains(lower, "click ") ||
		strings.Contains(lower, "classdef ") ||
		strings.Contains(lower, "style ")
}
