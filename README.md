# explain-html-gen

Generate beautiful, self-contained interactive HTML explanations from structured content. Designed for the `explain-diff-html` skill in VS Code Copilot Chat.

## Features

- **No external dependencies**: Fully self-contained HTML with inline CSS and JavaScript
- **Responsive design**: Works on desktop, tablet, and mobile
- **Interactive quiz**: Dynamically rendered with option randomization
- **Code blocks**: Proper whitespace preservation, syntax-aware formatting
- **Diagrams**: Validated Mermaid flowcharts rendered offline in tabbed panels
- **Accessible**: Focus states, semantic HTML, sufficient color contrast
- **Offline-first**: No CDNs, network calls, or external fonts

## Installation

### From source

```bash
cd /home/ben/Code/explain-html-gen
go build -o explain-html-gen ./cmd/explain-html-gen/main.go
```

This creates a standalone `explain-html-gen` binary.

### Add to PATH

```bash
sudo mv explain-html-gen /usr/local/bin/
```

## Usage

### From stdin

```bash
cat content.json | explain-html-gen
```

### From file

```bash
explain-html-gen --input content.json
```

### With custom output

```bash
explain-html-gen --input content.json --output /path/to/explanation.html
```

## Input Schema

The JSON input follows this structure:

```json
{
  "title": "string",
  "slug": "string",
  "summary": "string",
  "date": "YYYY-MM-DD",
  "author": "string",
  "background": {
    "intro": "string",
    "mental_model": "string (optional)",
    "components": [
      {
        "name": "string",
        "description": "string",
        "role": "string"
      }
    ],
    "prior": "string"
  },
  "intuition": {
    "intro": "string",
    "core_idea": "string",
    "old_behavior": "string (optional)",
    "new_behavior": "string",
    "trade_offs_edge_cases": "string"
  },
  "code": {
    "intro": "string",
    "subsections": [
      {
        "title": "string",
        "explanation": "string",
        "blocks": [
          {
            "language": "go|typescript|python|etc",
            "code": "string",
            "file": "string (optional)",
            "start_line": 10,
            "end_line": 20,
            "caption": "string (optional)"
          }
        ],
        "consequences": "string"
      }
    ]
  },
  "quiz": [
    {
      "question": "string",
      "options": ["option1", "option2", "option3", "option4"],
      "correct_idx": 0,
      "explanation": "string explaining the correct answer and common misconceptions"
    }
  ],
  "diagrams": [
    {
      "type": "mermaid",
      "title": "Architecture",
      "caption": "A high-level module flow.",
      "mermaid": "flowchart LR\nReader[Reader] --> HTML[Generated HTML]\nHTML --> Quiz[Interactive Quiz]"
    }
  ]
}
```

### Required fields

- `title` — Page title (e.g., "Explaining the API refactor")
- `slug` — URL-safe slug for filename (e.g., "api-refactor")
- `summary` — 1–2 sentence summary
- `background` — Background section with components and prior behavior
- `intuition` — Intuition section with core idea and trade-offs
- `code` — Code section with subsections and code blocks
- `quiz` — Exactly 5 quiz questions with 4 options each

### Optional fields

- `date` — Defaults to today in YYYY-MM-DD format
- `author` — Author attribution
- `diagrams` — Optional tabbed diagrams. `type: "mermaid"` supports validated Mermaid flowcharts using `graph` or `flowchart` with directions `TD`, `TB`, `BT`, `LR`, or `RL`, labeled nodes, and labeled edges. Unsupported Mermaid fails input validation before HTML generation.

## Examples

See [`examples/`](examples/) for complete JSON inputs and generated HTML files.

Quick example:

```json
{
  "title": "Understanding Request Routing",
  "slug": "request-routing",
  "summary": "Learn how the new routing system dispatches requests to handlers.",
  "background": {
    "intro": "Before diving into the change, let's understand the request flow.",
    "components": [
      {
        "name": "Router",
        "role": "Maps incoming requests to handlers",
        "description": "Central dispatcher for HTTP requests"
      }
    ],
    "prior": "Previously, routing was handled by a flat switch statement."
  },
  "intuition": {
    "intro": "The core insight is that routing should be composable.",
    "core_idea": "Routes are organized hierarchically, allowing reusable sub-routers.",
    "new_behavior": "Routes can now be nested and shared across the application.",
    "trade_offs_edge_cases": "Trade-off: slightly more complex setup for much greater flexibility."
  },
  "code": {
    "intro": "Here's how the implementation achieves this.",
    "subsections": [
      {
        "title": "Router Creation",
        "explanation": "The Router struct now uses a tree of route handlers.",
        "blocks": [
          {
            "language": "go",
            "code": "type Router struct {\n  routes map[string]Handler\n  children []*Router\n}",
            "file": "router.go",
            "start_line": 10,
            "end_line": 14,
            "caption": "Router structure now supports nesting"
          }
        ],
        "consequences": "This allows dynamic route registration and composition."
      }
    ]
  },
  "quiz": [
    {
      "question": "What is the primary benefit of the new routing system?",
      "options": [
        "Faster request handling",
        "Composable and nested route hierarchies",
        "Reduced memory usage",
        "Simpler global routing"
      ],
      "correct_idx": 1,
      "explanation": "The new system enables composable, reusable route hierarchies. While performance and memory are benefits, the core design goal is compositional structure."
    },
    {
      "question": "How are sub-routes registered in the new system?",
      "options": [
        "Via a flat global registry",
        "Using environment variables",
        "Through hierarchical Router nesting",
        "By modifying routing tables at runtime"
      ],
      "correct_idx": 2,
      "explanation": "Routes are now nested hierarchically, allowing sub-routers to be attached to parent routers, enabling composition and reusability."
    },
    {
      "question": "What trade-off is introduced with the new design?",
      "options": [
        "Increased latency for request processing",
        "Slightly more complex setup for greater flexibility",
        "Loss of request type safety",
        "Inability to use middleware"
      ],
      "correct_idx": 1,
      "explanation": "The hierarchical structure adds setup complexity, but provides significant gains in flexibility and code organization."
    },
    {
      "question": "In the old system, how were routes handled?",
      "options": [
        "Through composable hierarchies",
        "Using environment-based configuration",
        "Via a flat switch statement",
        "Through a global route registry"
      ],
      "correct_idx": 2,
      "explanation": "The previous implementation used a flat switch statement, which became difficult to maintain and extend."
    },
    {
      "question": "What data structure is used to store routes in the new implementation?",
      "options": [
        "A linked list",
        "A tree of nested Router instances",
        "A flat map with string keys only",
        "A priority queue"
      ],
      "correct_idx": 1,
      "explanation": "The new Router struct contains both a map for direct route lookup and a children field for nested Router instances, forming a tree structure."
    }
  ]
}
```

## Output Validation

The tool validates:

- **HTML structure**: DOCTYPE, html/head/body tags present
- **No external dependencies**: No CDNs, external fonts, or network calls
- **Code block whitespace**: CSS includes `white-space: pre` rule
- **Quiz interactivity**: JavaScript and quiz-card elements present
- **Required sections**: Background, Intuition, Code, Quiz all present
- **Mermaid input**: Mermaid diagram source is validated before rendering

If validation fails, the tool exits with status 1 and explains the problem.

For skill integration, invalid Mermaid can be repaired by catching stderr and passing
the exact validation error plus the original Mermaid source back to the language
model. For example: "The generator rejected this Mermaid diagram with
`unsupported mermaid syntax on line 2`; rewrite it using the supported flowchart
subset."

## Integration with explain-diff-html Skill

In the LM agent skill workflow:

1. **Gather content**: Explore codebase, build narrative sections, design quiz questions
2. **Output JSON**: Construct GeneratorInput struct and serialize to JSON
3. **Call generator**: Invoke `explain-html-gen --stdin < content.json`
4. **Use result**: Return the file:// URL from stdout

Example in skill workflow:

```
# After gathering narrative and quiz content...
json_content=$(cat <<'EOF'
{
  "title": "...",
  ...
}
EOF
)

result=$(echo "$json_content" | explain-html-gen)
echo "✓ Generated explanation: $result"
```

## Development

### Building

```bash
cd /home/ben/Code/explain-html-gen
go build -o explain-html-gen ./cmd/explain-html-gen/main.go
```

### Testing

```bash
go test ./...
```

### File structure

```
explain-html-gen/
├── cmd/
│   └── explain-html-gen/
│       └── main.go          # CLI entry point
├── pkg/
│   ├── generator/
│   │   ├── builder.go       # HTML builder
│   │   ├── quiz.go          # Quiz rendering & randomization (future)
│   │   └── diagrams.go      # Diagram patterns (future)
│   ├── schema/
│   │   ├── input.go         # JSON schema structs
│   │   └── error.go         # Validation errors
│   └── validator/
│       └── validator.go     # Output validation
├── examples/                # Example JSON inputs
├── go.mod
└── README.md
```

## License

This tool is part of the explain-diff-html skill and follows the same licensing.
