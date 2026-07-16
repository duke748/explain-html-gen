package schema

import "github.com/benthompson/explain-html-gen/pkg/mermaid"

// GeneratorInput defines the complete content structure for HTML generation.
// The LM agent populates this and passes it as JSON.
type GeneratorInput struct {
	// Metadata
	Title   string `json:"title"`   // Page title (e.g., "Explaining the API refactor")
	Slug    string `json:"slug"`    // URL-safe slug for filename (e.g., "api-refactor")
	Summary string `json:"summary"` // Short 1-2 sentence summary
	Date    string `json:"date"`    // YYYY-MM-DD format; if empty, uses today
	Author  string `json:"author"`  // Optional author attribution

	// Narrative sections
	Background BackgroundSection `json:"background"`
	Intuition  IntuitiveSection  `json:"intuition"`
	Code       CodeSection       `json:"code"`

	// Quiz: exactly 5 questions
	Quiz []QuizQuestion `json:"quiz"`

	// Optional diagrams and examples
	Diagrams []Diagram `json:"diagrams,omitempty"`
}

// BackgroundSection explains the system needed for understanding the change.
type BackgroundSection struct {
	Intro       string      `json:"intro"`                  // Opening paragraph(s)
	MentalModel *string     `json:"mental_model,omitempty"` // Optional beginner-friendly model
	Components  []Component `json:"components"`             // Key components involved
	Prior       string      `json:"prior"`                  // Prior behavior/constraints
}

// Component describes a system piece relevant to the change.
type Component struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Role        string `json:"role"` // e.g., "handles request routing"
}

// IntuitiveSection explains the core idea before implementation.
type IntuitiveSection struct {
	Intro              string  `json:"intro"`
	CoreIdea           string  `json:"core_idea"`
	OldBehavior        *string `json:"old_behavior,omitempty"`
	NewBehavior        string  `json:"new_behavior"`
	TradeOffsEdgeCases string  `json:"trade_offs_edge_cases"` // Edge cases & trade-offs
}

// CodeSection walks through implementation changes.
type CodeSection struct {
	Intro       string           `json:"intro"`
	Subsections []CodeSubsection `json:"subsections"` // Ordered by execution/dependency flow
}

// CodeSubsection groups related code changes.
type CodeSubsection struct {
	Title       string `json:"title"` // e.g., "Request parsing"
	Explanation string `json:"explanation"`
	// Code blocks with file references
	Blocks       []CodeBlock `json:"blocks"`
	Consequences string      `json:"consequences"` // Observable consequences of changes
}

// CodeBlock represents a code snippet with context.
type CodeBlock struct {
	Language  string `json:"language"`             // "go", "typescript", "python", etc.
	Code      string `json:"code"`                 // Actual code (will be escaped for HTML)
	File      string `json:"file"`                 // Optional: file path reference
	StartLine int    `json:"start_line,omitempty"` // Optional: line reference
	EndLine   int    `json:"end_line,omitempty"`   // Optional: line reference
	Caption   string `json:"caption,omitempty"`    // Optional: explanation
}

// QuizQuestion represents a single quiz question.
// Exactly 5 are required in GeneratorInput.Quiz.
type QuizQuestion struct {
	Question    string   `json:"question"`    // The question text
	Options     []string `json:"options"`     // 4 options (answer varies by question)
	CorrectIdx  int      `json:"correct_idx"` // Index of correct answer (0-3)
	Explanation string   `json:"explanation"` // Explanation of correct answer and common misconceptions
}

// Diagram represents a diagram or visualization.
// Type determines rendering:
// - "flow": request/data/control flow with labeled nodes and arrows
// - "before_after": side-by-side panels showing behavior change
// - "component_cards": labeled system boundaries
// - "table": mapping/invariant/toy data table
// - "mermaid": validated Mermaid flowchart source rendered as an offline diagram
type Diagram struct {
	Type    string `json:"type"` // "flow", "before_after", "component_cards", "table"
	Title   string `json:"title"`
	Caption string `json:"caption,omitempty"` // Text explanation alongside diagram

	// Type-specific content
	Mermaid      string          `json:"mermaid,omitempty"` // Mermaid flowchart source for type "mermaid"
	FlowNodes    []FlowNode      `json:"flow_nodes,omitempty"`
	FlowEdges    []FlowEdge      `json:"flow_edges,omitempty"`
	BeforePanel  PanelContent    `json:"before_panel,omitempty"`
	AfterPanel   PanelContent    `json:"after_panel,omitempty"`
	Components   []ComponentCard `json:"components,omitempty"`
	TableHeaders []string        `json:"table_headers,omitempty"`
	TableRows    [][]string      `json:"table_rows,omitempty"`
}

// FlowNode represents a node in a flow diagram.
type FlowNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value,omitempty"` // Example value flowing through node
}

// FlowEdge represents an arrow/connection in a flow diagram.
type FlowEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"` // Edge annotation
}

// PanelContent represents before/after panel content.
type PanelContent struct {
	Title string `json:"title"`
	Text  string `json:"text"`           // Main explanation
	Code  string `json:"code,omitempty"` // Optional code block
}

// ComponentCard describes a system boundary.
type ComponentCard struct {
	Name     string   `json:"name"`
	Border   string   `json:"border"`             // "solid", "dashed", or color
	Children []string `json:"children,omitempty"` // Component IDs contained within
}

// ValidateInput performs basic structural validation on the input.
// Returns an error if critical fields are missing or malformed.
func (g *GeneratorInput) Validate() error {
	if g.Title == "" {
		return NewValidationError("title is required")
	}
	if g.Slug == "" {
		return NewValidationError("slug is required")
	}
	if g.Summary == "" {
		return NewValidationError("summary is required")
	}
	if len(g.Quiz) != 5 {
		return NewValidationError("exactly 5 quiz questions are required")
	}
	for i, q := range g.Quiz {
		if q.Question == "" {
			return NewValidationError("quiz question %d has empty text", i+1)
		}
		if len(q.Options) != 4 {
			return NewValidationError("quiz question %d must have exactly 4 options", i+1)
		}
		if q.CorrectIdx < 0 || q.CorrectIdx > 3 {
			return NewValidationError("quiz question %d has invalid correct_idx", i+1)
		}
		if q.Explanation == "" {
			return NewValidationError("quiz question %d has empty explanation", i+1)
		}
	}
	for i, d := range g.Diagrams {
		if d.Type == "" {
			return NewValidationError("diagram %d has empty type", i+1)
		}
		if d.Title == "" {
			return NewValidationError("diagram %d has empty title", i+1)
		}
		if d.Type == "mermaid" {
			if d.Mermaid == "" {
				return NewValidationError("diagram %d is type mermaid but has empty mermaid source", i+1)
			}
			if err := mermaid.Validate(d.Mermaid); err != nil {
				return NewValidationError("diagram %d has invalid mermaid source: %v", i+1, err)
			}
		}
	}
	return nil
}
