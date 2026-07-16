package generator

import (
	"fmt"
	"html"
	"math/rand"
	"strings"
	"time"

	"github.com/benthompson/explain-html-gen/pkg/mermaid"
	"github.com/benthompson/explain-html-gen/pkg/schema"
)

// HTMLBuilder accumulates content and generates a complete HTML document.
type HTMLBuilder struct {
	input      *schema.GeneratorInput
	sections   []string // Accumulates HTML sections
	javaScript string   // JavaScript for quiz interactivity
}

// NewHTMLBuilder creates a new builder for the given input.
func NewHTMLBuilder(input *schema.GeneratorInput) *HTMLBuilder {
	return &HTMLBuilder{
		input:    input,
		sections: make([]string, 0),
	}
}

// Build generates the complete HTML document.
func (b *HTMLBuilder) Build() (string, error) {
	b.sections = make([]string, 0)

	// Generate each section
	b.addHead()
	b.addBodyStart()
	b.addHeader()
	b.addTableOfContents()
	b.addBackgroundSection()
	b.addDiagramsSection()
	b.addIntuitionSection()
	b.addCodeSection()
	b.addQuizSection()
	b.addBodyEnd()

	return strings.Join(b.sections, "\n"), nil
}

// addHead generates the HTML head with inline CSS.
func (b *HTMLBuilder) addHead() {
	date := b.input.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	css := b.generateCSS()
	js := b.generateJavaScript()

	head := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta name="description" content="%s">
	<title>%s</title>
	<style>
%s
	</style>
</head>`, html.EscapeString(b.input.Summary), html.EscapeString(b.input.Title), css)

	b.sections = append(b.sections, head)
	// Store JavaScript for later use
	b.javaScript = js
}

func (b *HTMLBuilder) addBodyStart() {
	b.sections = append(b.sections, "<body>")
	b.sections = append(b.sections, "<div class=\"container\">")
}

func (b *HTMLBuilder) addBodyEnd() {
	b.sections = append(b.sections, "</div>") // .container
	b.sections = append(b.sections, fmt.Sprintf("<script>\n%s\n</script>", b.javaScript))
	b.sections = append(b.sections, "</body>\n</html>")
}

func (b *HTMLBuilder) addHeader() {
	date := b.input.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	author := ""
	if b.input.Author != "" {
		author = fmt.Sprintf(`<p class="author">by %s</p>`, html.EscapeString(b.input.Author))
	}

	header := fmt.Sprintf(`<header>
	<h1>%s</h1>
	<p class="summary">%s</p>
	<p class="date">%s</p>
	%s
</header>`, html.EscapeString(b.input.Title), html.EscapeString(b.input.Summary), html.EscapeString(date), author)

	b.sections = append(b.sections, header)
}

func (b *HTMLBuilder) addTableOfContents() {
	toc := `<nav class="toc">
	<h2>Contents</h2>
	<ul>
		<li><a href="#background">Background</a></li>
`
	if len(b.input.Diagrams) > 0 {
		toc += `		<li><a href="#diagrams">Diagrams</a></li>
`
	}
	toc += `		<li><a href="#intuition">Intuition</a></li>
		<li><a href="#code">Code</a></li>
		<li><a href="#quiz">Quiz</a></li>
	</ul>
</nav>`

	b.sections = append(b.sections, toc)
}

func (b *HTMLBuilder) addBackgroundSection() {
	bg := b.input.Background
	section := fmt.Sprintf(`<section id="background">
	<h2>Background</h2>
	<div class="section-content">
		%s
`, strings.TrimSpace(escapeAndFormatText(bg.Intro)))

	if bg.MentalModel != nil && *bg.MentalModel != "" {
		section += fmt.Sprintf(`		<div class="callout callout-definition">
			<strong>Mental Model:</strong>
			<p>%s</p>
		</div>
`, strings.TrimSpace(escapeAndFormatText(*bg.MentalModel)))
	}

	if len(bg.Components) > 0 {
		section += `		<div class="components">
`
		for _, comp := range bg.Components {
			section += fmt.Sprintf(`			<div class="component-card">
				<h4>%s</h4>
				<p><strong>Role:</strong> %s</p>
				<p>%s</p>
			</div>
`, html.EscapeString(comp.Name), html.EscapeString(comp.Role), escapeAndFormatText(comp.Description))
		}
		section += `		</div>
`
	}

	section += fmt.Sprintf(`		<div class="callout callout-important">
			<strong>Prior Behavior:</strong>
			<p>%s</p>
		</div>
	</div>
</section>`, escapeAndFormatText(bg.Prior))

	b.sections = append(b.sections, section)
}

func (b *HTMLBuilder) addDiagramsSection() {
	if len(b.input.Diagrams) == 0 {
		return
	}

	section := `<section id="diagrams">
	<h2>Diagrams</h2>
	<div class="section-content">
		<div class="diagram-tabs" data-diagram-tabs>
			<div class="tab-list" role="tablist" aria-label="Diagrams">
`

	panels := ""
	for i, diagram := range b.input.Diagrams {
		tabID := fmt.Sprintf("diagram-tab-%d", i)
		panelID := fmt.Sprintf("diagram-panel-%d", i)
		selected := ""
		hidden := " hidden"
		tabIndex := "-1"
		if i == 0 {
			selected = ` aria-selected="true"`
			hidden = ""
			tabIndex = "0"
		} else {
			selected = ` aria-selected="false"`
		}

		section += fmt.Sprintf(`				<button id="%s" class="tab-button" role="tab" type="button" aria-controls="%s"%s tabindex="%s">%s</button>
`, tabID, panelID, selected, tabIndex, html.EscapeString(diagram.Title))

		panels += fmt.Sprintf(`			<div id="%s" class="tab-panel" role="tabpanel" aria-labelledby="%s"%s>
%s			</div>
`, panelID, tabID, hidden, b.formatDiagram(diagram))
	}

	section += `			</div>
`
	section += panels
	section += `		</div>
	</div>
</section>`

	b.sections = append(b.sections, section)
}

func (b *HTMLBuilder) formatDiagram(diagram schema.Diagram) string {
	switch diagram.Type {
	case "mermaid":
		return b.formatMermaidDiagram(diagram)
	default:
		return fmt.Sprintf(`				<div class="callout callout-important">
					<strong>Unsupported diagram type:</strong>
					<p>%s</p>
				</div>
`, html.EscapeString(diagram.Type))
	}
}

func (b *HTMLBuilder) formatMermaidDiagram(diagram schema.Diagram) string {
	parsed, err := mermaid.Parse(diagram.Mermaid)
	if err != nil {
		return fmt.Sprintf(`				<div class="callout callout-important">
					<strong>Invalid Mermaid:</strong>
					<p>%s</p>
				</div>
`, html.EscapeString(err.Error()))
	}

	orientationClass := "diagram-vertical"
	if parsed.Direction == "LR" || parsed.Direction == "RL" {
		orientationClass = "diagram-horizontal"
	}

	var out strings.Builder
	out.WriteString(fmt.Sprintf(`				<figure class="mermaid-diagram %s">
					<figcaption>
						<strong>%s</strong>
`, orientationClass, html.EscapeString(diagram.Title)))
	if diagram.Caption != "" {
		out.WriteString(fmt.Sprintf(`						<span>%s</span>
`, escapeAndFormatText(diagram.Caption)))
	}
	out.WriteString(`					</figcaption>
					<div class="mermaid-rendered" aria-label="Architecture diagram rendered from Mermaid source">
`)

	nodeLabels := make(map[string]string)
	for _, node := range parsed.Nodes {
		nodeLabels[node.ID] = node.Label
	}

	for i, edge := range parsed.Edges {
		out.WriteString(fmt.Sprintf(`						<div class="mermaid-edge">
							<div class="mermaid-node">
								<span class="node-id">%s</span>
								<span class="node-label">%s</span>
							</div>
`, html.EscapeString(edge.From), html.EscapeString(nodeLabels[edge.From])))
		if edge.Label != "" {
			out.WriteString(fmt.Sprintf(`							<div class="mermaid-arrow" aria-label="%s">&rarr; <span>%s</span></div>
`, html.EscapeString(edge.Label), html.EscapeString(edge.Label)))
		} else {
			out.WriteString(`							<div class="mermaid-arrow" aria-hidden="true">&rarr;</div>
`)
		}
		out.WriteString(fmt.Sprintf(`							<div class="mermaid-node">
								<span class="node-id">%s</span>
								<span class="node-label">%s</span>
							</div>
						</div>
`, html.EscapeString(edge.To), html.EscapeString(nodeLabels[edge.To])))
		if i < len(parsed.Edges)-1 {
			out.WriteString(`						<div class="mermaid-edge-separator" aria-hidden="true"></div>
`)
		}
	}

	out.WriteString(fmt.Sprintf(`					</div>
					<details class="mermaid-source">
						<summary>Mermaid source</summary>
						<pre><code style="white-space: pre-wrap; white-space: pre;">%s</code></pre>
					</details>
				</figure>
`, html.EscapeString(diagram.Mermaid)))

	return out.String()
}

func (b *HTMLBuilder) addIntuitionSection() {
	intuition := b.input.Intuition
	section := fmt.Sprintf(`<section id="intuition">
	<h2>Intuition</h2>
	<div class="section-content">
		%s
`, escapeAndFormatText(intuition.Intro))

	section += fmt.Sprintf(`		<div class="callout callout-definition">
			<strong>Core Idea:</strong>
			<p>%s</p>
		</div>
`, escapeAndFormatText(intuition.CoreIdea))

	if intuition.OldBehavior != nil && *intuition.OldBehavior != "" {
		section += fmt.Sprintf(`		<div class="before-after">
			<div class="panel">
				<h4>Old Behavior</h4>
				<p>%s</p>
			</div>
`, escapeAndFormatText(*intuition.OldBehavior))

		section += fmt.Sprintf(`			<div class="panel">
				<h4>New Behavior</h4>
				<p>%s</p>
			</div>
		</div>
`, escapeAndFormatText(intuition.NewBehavior))
	} else {
		section += fmt.Sprintf(`		<p><strong>New Behavior:</strong> %s</p>
`, escapeAndFormatText(intuition.NewBehavior))
	}

	section += fmt.Sprintf(`		<div class="callout callout-important">
			<strong>Trade-offs &amp; Edge Cases:</strong>
			<p>%s</p>
		</div>
	</div>
</section>`, escapeAndFormatText(intuition.TradeOffsEdgeCases))

	b.sections = append(b.sections, section)
}

func (b *HTMLBuilder) addCodeSection() {
	code := b.input.Code
	section := fmt.Sprintf(`<section id="code">
	<h2>Code</h2>
	<div class="section-content">
		%s
`, escapeAndFormatText(code.Intro))

	for _, subsection := range code.Subsections {
		section += fmt.Sprintf(`		<h3>%s</h3>
`, html.EscapeString(subsection.Title))
		section += fmt.Sprintf(`		<p>%s</p>
`, escapeAndFormatText(subsection.Explanation))

		for _, block := range subsection.Blocks {
			section += b.formatCodeBlock(block)
		}

		section += fmt.Sprintf(`		<p><em>%s</em></p>
`, escapeAndFormatText(subsection.Consequences))
	}

	section += `	</div>
</section>`

	b.sections = append(b.sections, section)
}

func (b *HTMLBuilder) formatCodeBlock(block schema.CodeBlock) string {
	var header string
	if block.File != "" {
		if block.StartLine > 0 {
			header = fmt.Sprintf(`<div class="code-header">%s (lines %d–%d)</div>
`, html.EscapeString(block.File), block.StartLine, block.EndLine)
		} else {
			header = fmt.Sprintf(`<div class="code-header">%s</div>
`, html.EscapeString(block.File))
		}
	}

	code := fmt.Sprintf(`<pre><code class="language-%s" style="white-space: pre-wrap; white-space: pre;">%s</code></pre>
`, html.EscapeString(block.Language), html.EscapeString(block.Code))

	caption := ""
	if block.Caption != "" {
		caption = fmt.Sprintf(`<p class="code-caption">%s</p>
`, escapeAndFormatText(block.Caption))
	}

	return fmt.Sprintf(`		<div class="code-block">
%s%s%s		</div>
`, header, code, caption)
}

func (b *HTMLBuilder) addQuizSection() {
	section := `<section id="quiz">
	<h2>Quiz</h2>
	<p class="quiz-intro">Test your understanding with these five questions. Click an option to see the answer and explanation.</p>
	<div class="quiz-container">
`
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i, q := range b.input.Quiz {
		section += b.formatQuizCard(i, q, rng)
	}

	section += `	</div>
</section>`

	b.sections = append(b.sections, section)
}

// shuffleOptions randomizes the order of quiz options using Fisher-Yates shuffle.
// Returns the shuffled options and the updated index of the correct answer.
func shuffleOptions(options []string, correctIdx int, rng *rand.Rand) ([]string, int) {
	if len(options) <= 1 {
		return options, correctIdx
	}

	// Track original indices
	type indexedOption struct {
		value   string
		origIdx int
	}
	indexed := make([]indexedOption, len(options))
	for i, v := range options {
		indexed[i] = indexedOption{value: v, origIdx: i}
	}

	// Fisher-Yates shuffle
	for i := len(indexed) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		indexed[i], indexed[j] = indexed[j], indexed[i]
	}

	// Extract shuffled options and track new correct index
	shuffled := make([]string, len(indexed))
	newCorrectIdx := 0
	for i, opt := range indexed {
		shuffled[i] = opt.value
		if opt.origIdx == correctIdx {
			newCorrectIdx = i
		}
	}

	return shuffled, newCorrectIdx
}

func (b *HTMLBuilder) formatQuizCard(idx int, q schema.QuizQuestion, rng *rand.Rand) string {
	// Build the quiz card as a template that will be populated by JS
	cardID := fmt.Sprintf("quiz-card-%d", idx)

	card := fmt.Sprintf(`		<div class="quiz-card" id="%s">
			<div class="quiz-question">
				<p><strong>Question %d:</strong> %s</p>
			</div>
			<div class="quiz-options">
`, cardID, idx+1, escapeAndFormatText(q.Question))

	// Randomize option order for this question.
	randomized, newCorrectIdx := shuffleOptions(q.Options, q.CorrectIdx, rng)

	for i, opt := range randomized {
		card += fmt.Sprintf(`				<label class="quiz-option" data-index="%d">
					<input type="radio" name="%s" value="%d" aria-label="Option %d">
					<span class="option-text">%s</span>
				</label>
`, i, cardID, i, i+1, escapeAndFormatText(opt))
	}

	card += fmt.Sprintf(`			</div>
			<div class="quiz-feedback" data-correct-idx="%d" data-explanation="%s" style="display: none;">
			</div>
		</div>
`, newCorrectIdx, html.EscapeString(q.Explanation))

	return card
}

func (b *HTMLBuilder) generateCSS() string {
	return `
/* Reset and base styles */
* {
	margin: 0;
	padding: 0;
	box-sizing: border-box;
}

html {
	scroll-behavior: smooth;
}

body {
	font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
	line-height: 1.6;
	color: #333;
	background: #f9f9f9;
}

.container {
	max-width: 900px;
	margin: 0 auto;
	padding: 20px;
	background: white;
	box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* Header and metadata */
header {
	border-bottom: 3px solid #0078d4;
	padding-bottom: 20px;
	margin-bottom: 40px;
}

header h1 {
	font-size: 2.5rem;
	margin-bottom: 10px;
	color: #0078d4;
}

header .summary {
	font-size: 1.2rem;
	color: #666;
	margin-bottom: 10px;
	font-weight: 500;
}

header .date, header .author {
	font-size: 0.9rem;
	color: #999;
}

/* Table of Contents */
nav.toc {
	background: #f0f7ff;
	border-left: 4px solid #0078d4;
	padding: 20px;
	margin: 30px 0;
	border-radius: 4px;
}

nav.toc h2 {
	font-size: 1.2rem;
	margin-bottom: 15px;
	color: #0078d4;
}

nav.toc ul {
	list-style: none;
}

nav.toc li {
	margin: 8px 0;
}

nav.toc a {
	color: #0078d4;
	text-decoration: none;
	border-bottom: 1px solid transparent;
	transition: border-color 0.2s;
}

nav.toc a:hover {
	border-bottom-color: #0078d4;
}

/* Sections */
section {
	margin: 50px 0;
	scroll-margin-top: 20px;
}

section h2 {
	font-size: 2rem;
	color: #0078d4;
	margin-bottom: 20px;
	border-bottom: 2px solid #e0e0e0;
	padding-bottom: 10px;
}

section h3 {
	font-size: 1.4rem;
	color: #333;
	margin: 30px 0 15px;
}

.section-content {
	line-height: 1.8;
	color: #555;
}

.section-content p {
	margin-bottom: 15px;
}

/* Callouts */
.callout {
	border-left: 4px solid #666;
	padding: 15px 20px;
	margin: 20px 0;
	background: #f5f5f5;
	border-radius: 4px;
}

.callout-definition {
	border-left-color: #0078d4;
	background: #f0f7ff;
}

.callout-important {
	border-left-color: #ff8c00;
	background: #fff8f0;
}

.callout strong {
	color: #0078d4;
}

.callout-important strong {
	color: #ff8c00;
}

/* Components */
.components {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
	gap: 15px;
	margin: 20px 0;
}

.component-card {
	border: 1px solid #ddd;
	border-radius: 6px;
	padding: 15px;
	background: #fafafa;
}

.component-card h4 {
	color: #0078d4;
	margin-bottom: 10px;
	font-size: 1.1rem;
}

.component-card p {
	font-size: 0.95rem;
	margin-bottom: 8px;
}

/* Before/After panels */
.before-after {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: 20px;
	margin: 20px 0;
}

.before-after .panel {
	border: 1px solid #ddd;
	border-radius: 6px;
	padding: 15px;
	background: #fafafa;
}

.before-after .panel h4 {
	color: #0078d4;
	margin-bottom: 10px;
	font-weight: 600;
}

@media (max-width: 600px) {
	.before-after {
		grid-template-columns: 1fr;
	}
}

/* Code blocks */
.code-block {
	margin: 20px 0;
	border: 1px solid #ddd;
	border-radius: 6px;
	overflow: hidden;
	background: #f8f8f8;
}

.code-header {
	background: #e8e8e8;
	padding: 10px 15px;
	font-size: 0.9rem;
	font-weight: 600;
	color: #333;
	border-bottom: 1px solid #ddd;
}

.code-block pre {
	margin: 0;
	padding: 15px;
	overflow-x: auto;
}

.code-block code {
	font-family: "Monaco", "Menlo", "Consolas", monospace;
	font-size: 0.9rem;
	line-height: 1.5;
	color: #333;
	white-space: pre;
	white-space: pre-wrap;
}

.code-caption {
	padding: 0 15px 10px;
	font-size: 0.9rem;
	color: #666;
	font-style: italic;
	border-top: 1px solid #ddd;
}

/* Diagrams */
.diagram-tabs {
	margin: 20px 0;
}

.tab-list {
	display: flex;
	flex-wrap: wrap;
	gap: 8px;
	margin-bottom: 16px;
	border-bottom: 1px solid #ddd;
}

.tab-button {
	border: 1px solid #ddd;
	border-bottom: 0;
	border-radius: 6px 6px 0 0;
	padding: 10px 14px;
	background: #f8f8f8;
	color: #333;
	font: inherit;
	font-weight: 600;
	cursor: pointer;
}

.tab-button[aria-selected="true"] {
	background: white;
	color: #0078d4;
	border-color: #0078d4;
}

.tab-panel {
	border: 1px solid #ddd;
	border-radius: 0 6px 6px 6px;
	padding: 18px;
	background: white;
}

.mermaid-diagram {
	margin: 0;
}

.mermaid-diagram figcaption {
	display: grid;
	gap: 4px;
	margin-bottom: 16px;
	color: #333;
}

.mermaid-diagram figcaption span {
	color: #666;
	font-size: 0.95rem;
}

.mermaid-rendered {
	display: grid;
	gap: 10px;
	padding: 16px;
	border: 1px solid #d8e6f3;
	border-radius: 6px;
	background: #f7fbff;
	overflow-x: auto;
}

.mermaid-edge {
	display: grid;
	grid-template-columns: minmax(140px, 1fr) auto minmax(140px, 1fr);
	gap: 12px;
	align-items: center;
	min-width: 0;
}

.mermaid-node {
	display: grid;
	gap: 4px;
	border: 1px solid #9ec7ea;
	border-radius: 6px;
	padding: 12px;
	background: white;
	min-width: 0;
}

.node-id {
	color: #0078d4;
	font-size: 0.78rem;
	font-weight: 700;
	letter-spacing: 0;
	text-transform: uppercase;
}

.node-label {
	color: #333;
	font-weight: 600;
	overflow-wrap: anywhere;
}

.mermaid-arrow {
	display: grid;
	gap: 2px;
	justify-items: center;
	color: #0078d4;
	font-weight: 700;
	white-space: nowrap;
}

.mermaid-arrow span {
	max-width: 160px;
	color: #555;
	font-size: 0.82rem;
	font-weight: 600;
	overflow-wrap: anywhere;
	white-space: normal;
	text-align: center;
}

.mermaid-edge-separator {
	height: 1px;
	background: #d8e6f3;
}

.mermaid-source {
	margin-top: 14px;
}

.mermaid-source summary {
	cursor: pointer;
	color: #0078d4;
	font-weight: 600;
}

.mermaid-source pre {
	margin-top: 10px;
	padding: 12px;
	overflow-x: auto;
	border-radius: 6px;
	background: #f8f8f8;
}

/* Quiz */
#quiz {
	margin-top: 60px;
}

.quiz-intro {
	font-size: 1.1rem;
	margin-bottom: 20px;
	color: #666;
}

.quiz-container {
	display: grid;
	gap: 30px;
}

.quiz-card {
	border: 1px solid #ddd;
	border-radius: 8px;
	padding: 20px;
	background: #f9f9f9;
}

.quiz-card.answered-correct {
	border-color: #28a745;
	background: #f0fdf4;
}

.quiz-card.answered-incorrect {
	border-color: #dc3545;
	background: #fdf0f0;
}

.quiz-question {
	margin-bottom: 20px;
}

.quiz-question p {
	font-size: 1.1rem;
	line-height: 1.6;
	color: #333;
}

.quiz-options {
	display: grid;
	gap: 12px;
	margin: 20px 0;
}

.quiz-option {
	display: grid;
	grid-template-columns: auto 1fr;
	gap: 12px;
	align-items: start;
	padding: 12px;
	border: 2px solid #ddd;
	border-radius: 6px;
	cursor: pointer;
	transition: all 0.2s;
	background: white;
}

.quiz-option:hover {
	border-color: #0078d4;
	background: #f0f7ff;
}

.quiz-option input[type="radio"] {
	margin-top: 3px;
	cursor: pointer;
	accent-color: #0078d4;
}

.quiz-option .option-text {
	word-break: break-word;
	color: #333;
}

.quiz-option input:disabled {
	cursor: not-allowed;
}

.quiz-option.selected {
	border-width: 2px;
}

.quiz-option.selected.correct {
	border-color: #28a745;
	background: #f0fdf4;
}

.quiz-option.selected.incorrect {
	border-color: #dc3545;
	background: #fdf0f0;
}

.quiz-feedback {
	margin-top: 20px;
	padding: 15px;
	border-radius: 6px;
	border-left: 4px solid #0078d4;
	background: #f0f7ff;
}

.quiz-feedback.correct {
	border-left-color: #28a745;
	background: #f0fdf4;
}

.quiz-feedback.incorrect {
	border-left-color: #dc3545;
	background: #fdf0f0;
}

.quiz-feedback-label {
	font-weight: 600;
	margin-bottom: 8px;
	color: #0078d4;
}

.quiz-feedback.correct .quiz-feedback-label {
	color: #28a745;
}

.quiz-feedback.incorrect .quiz-feedback-label {
	color: #dc3545;
}

.quiz-feedback p {
	margin: 8px 0;
	font-size: 0.95rem;
	color: #555;
}

/* Responsive design */
@media (max-width: 768px) {
	.container {
		padding: 15px;
	}

	header h1 {
		font-size: 1.8rem;
	}

	section h2 {
		font-size: 1.5rem;
	}

	section h3 {
		font-size: 1.2rem;
	}

	.components {
		grid-template-columns: 1fr;
	}

	.code-block code {
		font-size: 0.85rem;
	}

	.mermaid-edge {
		grid-template-columns: 1fr;
	}

	.mermaid-arrow {
		transform: rotate(90deg);
	}

	.mermaid-arrow span {
		transform: rotate(-90deg);
	}
}

@media (max-width: 480px) {
	.container {
		padding: 10px;
	}

	header h1 {
		font-size: 1.5rem;
	}

	header .summary {
		font-size: 1rem;
	}

	section h2 {
		font-size: 1.3rem;
	}

	nav.toc {
		padding: 15px;
	}

	.quiz-option {
		padding: 10px;
		font-size: 0.95rem;
	}
}

/* Focus states for accessibility */
input:focus-visible,
a:focus-visible,
button:focus-visible {
	outline: 2px solid #0078d4;
	outline-offset: 2px;
}

/* Print styles */
@media print {
	body {
		background: white;
	}

	.container {
		box-shadow: none;
		max-width: 100%;
	}

	section {
		page-break-inside: avoid;
	}
}`
}

func (b *HTMLBuilder) generateJavaScript() string {
	return `
document.addEventListener('DOMContentLoaded', function() {
	initDiagramTabs();
	initQuiz();
});

function initDiagramTabs() {
	const tabGroups = document.querySelectorAll('[data-diagram-tabs]');

	tabGroups.forEach(group => {
		const tabs = Array.from(group.querySelectorAll('[role="tab"]'));
		const panels = Array.from(group.querySelectorAll('[role="tabpanel"]'));

		function selectTab(tab) {
			tabs.forEach(item => {
				const selected = item === tab;
				item.setAttribute('aria-selected', selected ? 'true' : 'false');
				item.setAttribute('tabindex', selected ? '0' : '-1');
			});

			panels.forEach(panel => {
				panel.hidden = panel.id !== tab.getAttribute('aria-controls');
			});
		}

		tabs.forEach((tab, idx) => {
			tab.addEventListener('click', () => selectTab(tab));
			tab.addEventListener('keydown', event => {
				if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft') {
					return;
				}
				event.preventDefault();
				const delta = event.key === 'ArrowRight' ? 1 : -1;
				const next = tabs[(idx + delta + tabs.length) % tabs.length];
				selectTab(next);
				next.focus();
			});
		});
	});
}

// Quiz interaction logic
function initQuiz() {
	const quizCards = document.querySelectorAll('.quiz-card');

	quizCards.forEach(card => {
		const options = card.querySelectorAll('.quiz-option');
		const feedbackEl = card.querySelector('.quiz-feedback');
		const correctIdx = parseInt(feedbackEl.dataset.correctIdx);
		const explanation = feedbackEl.dataset.explanation;

		options.forEach((option, idx) => {
			const radio = option.querySelector('input[type="radio"]');
			radio.addEventListener('change', function() {
				handleQuizAnswer(card, options, idx, correctIdx, explanation);
			});
		});
	});
}

function handleQuizAnswer(card, options, selectedIdx, correctIdx, explanation) {
	const feedbackEl = card.querySelector('.quiz-feedback');
	const isCorrect = selectedIdx === correctIdx;

	// Disable all options
	options.forEach(opt => {
		opt.querySelector('input[type="radio"]').disabled = true;
	});

	// Mark selected option
	options[selectedIdx].classList.add('selected');
	if (isCorrect) {
		options[selectedIdx].classList.add('correct');
	} else {
		options[selectedIdx].classList.add('incorrect');
	}

	// Mark correct option if user was wrong
	if (!isCorrect) {
		options[correctIdx].classList.add('selected', 'correct');
	}

	// Show feedback
	feedbackEl.classList.add(isCorrect ? 'correct' : 'incorrect');
	feedbackEl.style.display = 'block';

	const label = isCorrect ? '✓ Correct!' : '✗ Incorrect';
	feedbackEl.innerHTML = '<div class="quiz-feedback-label">' + label + '</div><p>' + explanation + '</p>';

	// Mark card
	card.classList.add(isCorrect ? 'answered-correct' : 'answered-incorrect');
}
`
}

// Helper functions

func escapeAndFormatText(text string) string {
	// Escape HTML special characters
	escaped := html.EscapeString(text)
	// Convert simple line breaks (this is basic; enhance as needed)
	escaped = strings.ReplaceAll(escaped, "\n\n", "</p><p>")
	if !strings.HasPrefix(escaped, "<p>") {
		escaped = "<p>" + escaped
	}
	if !strings.HasSuffix(escaped, "</p>") {
		escaped = escaped + "</p>"
	}
	return escaped
}
