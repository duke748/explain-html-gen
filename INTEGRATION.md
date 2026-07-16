# explain-html-gen Integration Guide

## For LM Agent Developers (Copilot Skill Authors)

This guide explains how to integrate `explain-html-gen` into the `explain-diff-html` skill.

### Overview

The generator takes structured content (narrative sections, quiz questions, diagrams) as JSON input and produces a complete, self-contained interactive HTML file. This eliminates the need to manually write HTML scaffolding, CSS, and quiz logic in the skill workflow.

### Workflow

1. **Gather content** → Explore the codebase, build narrative sections, design quiz questions
2. **Construct JSON** → Create a `GeneratorInput` structure with all content
3. **Call generator** → Pass JSON to the CLI tool
4. **Return result** → Output the file:// URL to the user

### Example Skill Integration

In your `.instructions.md` or skill workflow:

```markdown
# explain-diff-html Skill Integration

1. Identify the change and explore the surrounding code (existing behavior)
2. Build the narrative:
   - Background: system context, prior behavior
   - Intuition: core idea, trade-offs
   - Code: implementation details with file references
   - Quiz: 5 questions testing understanding
3. Construct JSON:
   {
     "title": "...",
     "slug": "...",
     ...
   }
4. Call generator:
   result=$(echo "$json_content" | explain-html-gen)
5. Output result:
   echo "$result"
```

### Input Structure

See the main README for full schema. Key points:

- **Exactly 5 quiz questions** required (validates on input)
- **Options**: 4 per question, with `correct_idx` (0-3)
- **Explanations**: should explain the correct answer AND address misconceptions
- **Code blocks**: include file, line numbers, language, and caption

### Calling the Generator

#### From stdin (recommended for shell integration):

```bash
cat content.json | explain-html-gen
```

#### From file:

```bash
explain-html-gen --input content.json
```

#### With custom output path:

```bash
explain-html-gen --input content.json --output /path/to/my-explanation.html
```

### Output

On success, prints:
```
✓ Generated: file:///tmp/2026-07-16-explanation-my-slug.html
```

On failure, prints error details to stderr and exits with status 1.

### Validation Performed

The generator validates:

✓ **HTML structure** — valid DOCTYPE, html/head/body  
✓ **No external deps** — no CDNs, external fonts, network calls  
✓ **Code blocks** — CSS includes `white-space: pre`  
✓ **Quiz completeness** — exactly 5 questions, all required fields  
✓ **Required sections** — Background, Intuition, Code, Quiz present  

### Python Integration (if using LM agents in Python)

```python
import json
import subprocess

# Build content
content = {
    "title": "...",
    "slug": "...",
    # ... rest of schema
}

# Call generator
result = subprocess.run(
    ["explain-html-gen", "--stdin"],
    input=json.dumps(content),
    capture_output=True,
    text=True
)

if result.returncode != 0:
    print(f"Generator failed: {result.stderr}")
else:
    print(result.stdout)
```

### Node.js/JavaScript Integration

```javascript
const { execSync } = require('child_process');

const content = {
    title: "...",
    slug: "...",
    // ... rest of schema
};

try {
    const result = execSync('explain-html-gen', {
        input: JSON.stringify(content),
        encoding: 'utf-8'
    });
    console.log(result);
} catch (error) {
    console.error('Generator failed:', error.message);
}
```

### Building a Custom Example

1. Start with `examples/routing-example.json` as a template
2. Customize the sections with your content
3. Test locally:
   ```bash
   explain-html-gen --input my-example.json
   ```
4. Open the generated HTML in a browser to verify layout and quiz interactions

### Tips

- **Slugs**: Use lowercase, hyphens only (no spaces/underscores) — e.g., `api-refactor`, `database-migration`
- **Dates**: Optional; defaults to today. Format: YYYY-MM-DD
- **Quiz design**: See skill spec for best practices on option balance, distractors, and answer position randomization
- **Code blocks**: Preserve meaningful whitespace in code; the tool escapes HTML automatically
- **Line numbers**: Include start_line and end_line when pointing to specific code locations

### Troubleshooting

**Error: "exactly 5 quiz questions are required"**
- Count your quiz array; must have exactly 5 questions

**Error: "HTML contains external dependency"**
- Remove any CDN links, @import statements, or external font references

**Error: "JSON parsing failed"**
- Validate JSON syntax (check for trailing commas, unescaped quotes)
- Ensure file encoding is UTF-8

**HTML doesn't look right**
- Open the generated file in a browser to check CSS and layout
- Check that responsive CSS is applied on mobile (use browser DevTools)

**Quiz isn't interactive**
- Verify JavaScript is enabled in the output (check source with view-source)
- All data attributes must be properly set (data-correct-idx, data-explanation)

### Future Enhancements

Planned additions:

- [ ] Diagram patterns: flow diagrams, before/after panels, component hierarchies
- [ ] Syntax highlighting for code blocks
- [ ] Search/filter functionality for large explanations
- [ ] PDF export
- [ ] Collaborative editing API

## For Generator Developers

See the main `README.md` and package code for architecture details.

Key packages:

- `pkg/schema/input.go` — Input validation and schema
- `pkg/generator/builder.go` — HTML generation and CSS
- `pkg/validator/validator.go` — Output validation
- `cmd/explain-html-gen/main.go` — CLI

Extending the tool:

1. **Add diagram types**: Expand `schema.Diagram` type and `builder.formatDiagram()` function
2. **Customize CSS**: Edit `builder.generateCSS()`
3. **Modify validation rules**: Add checks to `validator.go`
4. **Implement CLI features**: Extend `main.go` flag parsing

## License

Same as the explain-diff-html skill.
