package validator

import (
	"errors"
	"regexp"
	"strings"
)

// HTMLValidator checks generated HTML for compliance with skill requirements.
type HTMLValidator struct {
	html string
}

// NewHTMLValidator creates a validator for the given HTML string.
func NewHTMLValidator(html string) *HTMLValidator {
	return &HTMLValidator{html: html}
}

// Validate performs comprehensive checks on the HTML.
// Returns an error if any check fails.
func (v *HTMLValidator) Validate() error {
	checks := []func() error{
		v.checkIsValidHTML,
		v.checkNoExternalDependencies,
		v.checkCodeBlocksPreserveWhitespace,
		v.checkQuizInteractivity,
		v.checkRequiredSections,
	}

	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

// checkIsValidHTML verifies the HTML has proper DOCTYPE and structure.
func (v *HTMLValidator) checkIsValidHTML() error {
	if !strings.Contains(v.html, "<!DOCTYPE html>") {
		return errors.New("HTML missing <!DOCTYPE html>")
	}
	if !strings.Contains(v.html, "<html") || !strings.Contains(v.html, "</html>") {
		return errors.New("HTML missing <html> tags")
	}
	if !strings.Contains(v.html, "<head") || !strings.Contains(v.html, "</head>") {
		return errors.New("HTML missing <head> section")
	}
	if !strings.Contains(v.html, "<body") || !strings.Contains(v.html, "</body>") {
		return errors.New("HTML missing <body> section")
	}
	return nil
}

// checkNoExternalDependencies verifies no external CDNs, fonts, or network resources.
func (v *HTMLValidator) checkNoExternalDependencies() error {
	forbidden := []string{
		"http://cdn.",
		"https://cdn.",
		"@import url",
		"@font-face",
		"<img src=\"http",
		"<img src=\"https",
		"<link href=\"http",
		"<link href=\"https",
		"<script src=\"http",
		"<script src=\"https",
	}

	for _, pattern := range forbidden {
		if strings.Contains(v.html, pattern) {
			return errors.New("HTML contains external dependency: " + pattern)
		}
	}
	return nil
}

// checkCodeBlocksPreserveWhitespace verifies <pre><code> blocks have explicit whitespace CSS.
func (v *HTMLValidator) checkCodeBlocksPreserveWhitespace() error {
	// Check that pre/code elements have white-space: pre or white-space: pre-wrap
	preRegex := regexp.MustCompile(`<pre[^>]*>`)
	matches := preRegex.FindAllStringIndex(v.html, -1)
	if len(matches) == 0 {
		// No code blocks is okay if there are no code samples
		return nil
	}

	// Check that CSS explicitly sets white-space: pre for code elements
	if !strings.Contains(v.html, "white-space: pre") {
		return errors.New("CSS missing 'white-space: pre' rule for code blocks")
	}

	return nil
}

// checkQuizInteractivity verifies quiz JavaScript is present and functional.
func (v *HTMLValidator) checkQuizInteractivity() error {
	if !strings.Contains(v.html, "<script>") || !strings.Contains(v.html, "</script>") {
		return errors.New("HTML missing <script> tag")
	}

	if !strings.Contains(v.html, "handleQuizAnswer") {
		return errors.New("JavaScript missing quiz answer handler")
	}

	if !strings.Contains(v.html, "quiz-card") {
		return errors.New("HTML missing quiz-card elements")
	}

	return nil
}

// checkRequiredSections verifies all required page sections are present.
func (v *HTMLValidator) checkRequiredSections() error {
	sections := map[string]string{
		"Background": `id="background"`,
		"Intuition":  `id="intuition"`,
		"Code":       `id="code"`,
		"Quiz":       `id="quiz"`,
	}

	for name, marker := range sections {
		if !strings.Contains(v.html, marker) {
			return errors.New("HTML missing " + name + " section")
		}
	}

	return nil
}

// CheckQuizBalance verifies that:
// 1. Exactly 5 quiz questions exist
// 2. Correct answers are balanced across positions
// Returns warnings (not errors) if balance is suboptimal.
func (v *HTMLValidator) CheckQuizBalance() []string {
	var warnings []string

	// Count quiz cards
	quizCardRegex := regexp.MustCompile(`id="quiz-card-\d+"`)
	matches := quizCardRegex.FindAllString(v.html, -1)
	if len(matches) != 5 {
		warnings = append(warnings, "Expected 5 quiz questions, found "+string(rune(len(matches))))
	}

	// Check that correct answers appear in different positions
	// This is a heuristic check; actual verification requires parsing data-correct-idx
	correctIdxRegex := regexp.MustCompile(`data-correct-idx="(\d)"`)
	idxMatches := correctIdxRegex.FindAllStringSubmatch(v.html, -1)
	if len(idxMatches) < 3 {
		warnings = append(warnings, "Could not verify correct answer position balance (found "+string(rune(len(idxMatches)))+" data-correct-idx attributes)")
	} else {
		// Simple check: collect the indices
		positions := make(map[string]int)
		for _, match := range idxMatches {
			positions[match[1]]++
		}
		// Ideally, positions should be fairly balanced (each position 0-3 appears roughly equally)
		maxCount := 0
		for _, count := range positions {
			if count > maxCount {
				maxCount = count
			}
		}
		if maxCount > 2 {
			warnings = append(warnings, "Correct answers appear to favor certain positions; ensure randomization is working")
		}
	}

	return warnings
}
