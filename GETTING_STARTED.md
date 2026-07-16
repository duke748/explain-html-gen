# Getting Started

## Quick Start

### 1. Build the binary

```bash
cd /home/ben/Code/explain-html-gen
go build -o bin/explain-html-gen ./cmd/explain-html-gen
```

### 2. Generate an explanation

```bash
# Using an example
./bin/explain-html-gen --input examples/routing-example.json

# Or from stdin
cat examples/routing-example.json | ./bin/explain-html-gen
```

### 3. Open the result

Look for the output path printed:
```
✓ Generated: file:///tmp/2026-07-16-explanation-request-routing.html
```

Open that path in a browser to see the interactive explanation with working quiz.

## For Skill Integration

### Python example

```python
import json
import subprocess

content = {
    "title": "My Explanation",
    "slug": "my-explanation",
    "summary": "A brief summary.",
    "background": { "intro": "...", "components": [...], "prior": "..." },
    "intuition": { "intro": "...", "core_idea": "...", ... },
    "code": { "intro": "...", "subsections": [...] },
    "quiz": [
        {
            "question": "...",
            "options": ["a", "b", "c", "d"],
            "correct_idx": 0,
            "explanation": "..."
        },
        # ... 4 more questions
    ]
}

result = subprocess.run(
    ["./bin/explain-html-gen"],
    input=json.dumps(content),
    capture_output=True,
    text=True
)

if result.returncode == 0:
    print(result.stdout)  # File path
else:
    print(f"Error: {result.stderr}")
```

### Bash example

```bash
json_content=$(cat <<'EOF'
{
  "title": "...",
  "slug": "...",
  ...
}
EOF
)

result=$(echo "$json_content" | ./bin/explain-html-gen)
echo "Generated: $result"
```

## File Structure

```
explain-html-gen/
├── bin/
│   └── explain-html-gen          # Compiled binary (run after `go build`)
├── cmd/
│   └── explain-html-gen/
│       └── main.go               # CLI entry point
├── pkg/
│   ├── generator/
│   │   └── builder.go            # HTML generation with CSS & JavaScript
│   ├── schema/
│   │   ├── input.go              # JSON schema & validation
│   │   └── error.go              # Error types
│   └── validator/
│       └── validator.go          # Output validation
├── examples/
│   └── routing-example.json      # Example JSON (tested)
├── go.mod                        # Go module definition
├── README.md                     # Full documentation
├── INTEGRATION.md                # Integration guide for skill authors
├── GETTING_STARTED.md            # This file
└── LICENSE                       # (to be created)
```

## Architecture

- **Input**: JSON following `schema.GeneratorInput` struct
- **Processing**: `HTMLBuilder` generates complete HTML with inline CSS & JavaScript
- **Validation**: `HTMLValidator` checks output completeness and constraints
- **Output**: Self-contained HTML file with no external dependencies

## Next Steps

1. **Integrate with skill**: See [INTEGRATION.md](INTEGRATION.md)
2. **Create custom examples**: Start with `examples/routing-example.json`
3. **Extend**: Add diagram patterns in `pkg/generator/diagrams.go` (stub ready)
4. **Test**: Run `go test ./...` (test framework ready for implementation)

## Troubleshooting

**"explain-html-gen: command not found"**
- Run `go build -o bin/explain-html-gen ./cmd/explain-html-gen` first
- Or use full path: `./bin/explain-html-gen`

**"exactly 5 quiz questions are required"**
- Your JSON must have exactly 5 questions in the `quiz` array

**"invalid JSON"**
- Check JSON syntax with `jq . < examples/routing-example.json`

**"HTML contains external dependency"**
- Remove any CDN links, @import statements, or external fonts from your content

## Support

See the full [README.md](README.md) for comprehensive documentation, [INTEGRATION.md](INTEGRATION.md) for skill integration details, and inline code comments for implementation specifics.
