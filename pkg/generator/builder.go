package generator

import (
	"fmt"
	"html"
	"strings"
	"time"

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
		<li><a href="#intuition">Intuition</a></li>
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

	for i, q := range b.input.Quiz {
		section += b.formatQuizCard(i, q)
	}

	section += `	</div>
</section>`

	b.sections = append(b.sections, section)
}

func (b *HTMLBuilder) formatQuizCard(idx int, q schema.QuizQuestion) string {
	// Build the quiz card as a template that will be populated by JS
	cardID := fmt.Sprintf("quiz-card-%d", idx)

	card := fmt.Sprintf(`		<div class="quiz-card" id="%s">
			<div class="quiz-question">
				<p><strong>Question %d:</strong> %s</p>
			</div>
			<div class="quiz-options">
`, cardID, idx+1, escapeAndFormatText(q.Question))

	// Options are stored in data attributes; JS will randomize and display
	for i, opt := range q.Options {
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
`, q.CorrectIdx, html.EscapeString(q.Explanation))

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
a:focus-visible {
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
// Quiz interaction logic
document.addEventListener('DOMContentLoaded', function() {
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
});

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
