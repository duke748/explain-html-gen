package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/benthompson/explain-html-gen/pkg/generator"
	"github.com/benthompson/explain-html-gen/pkg/schema"
	"github.com/benthompson/explain-html-gen/pkg/validator"
)

func main() {
	var (
		inputFile  = flag.String("input", "", "Path to JSON input file (or use --stdin)")
		outputPath = flag.String("output", "", "Path to output HTML file (default: /tmp/YYYY-MM-DD-explanation-<slug>.html)")
		stdin      = flag.Bool("stdin", false, "Read JSON from stdin instead of file")
		help       = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// Determine input source
	var inputData []byte
	var err error

	if *stdin {
		inputData, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	} else if *inputFile != "" {
		inputData, err = os.ReadFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: provide --input <file> or use --stdin\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse JSON
	var input schema.GeneratorInput
	if err := json.Unmarshal(inputData, &input); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON input: %v\n", err)
		os.Exit(1)
	}

	// Validate input
	if err := input.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Input validation failed: %v\n", err)
		os.Exit(1)
	}

	// Generate HTML
	builder := generator.NewHTMLBuilder(&input)
	html, err := builder.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating HTML: %v\n", err)
		os.Exit(1)
	}

	// Validate output
	v := validator.NewHTMLValidator(html)
	if err := v.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Output validation failed: %v\n", err)
		os.Exit(1)
	}

	// Warn about quiz balance issues
	warnings := v.CheckQuizBalance()
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
	}

	// Determine output path
	outPath := *outputPath
	if outPath == "" {
		date := input.Date
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		slug := input.Slug
		filename := fmt.Sprintf("%s-explanation-%s.html", date, slug)
		// Prefer ~/Downloads so snap-confined browsers (e.g. Firefox snap) can
		// access the file. Fall back to the user home dir, then /tmp.
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = ""
		}
		downloadsDir := filepath.Join(homeDir, "Downloads")
		if homeDir != "" {
			if info, statErr := os.Stat(downloadsDir); statErr == nil && info.IsDir() {
				outPath = filepath.Join(downloadsDir, filename)
			} else {
				outPath = filepath.Join(homeDir, filename)
			}
		} else {
			outPath = filepath.Join("/tmp", filename)
		}
	}

	// Ensure output directory exists
	outDir := filepath.Dir(outPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if err := os.WriteFile(outPath, []byte(html), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	// Print result
	absPath, err := filepath.Abs(outPath)
	if err != nil {
		absPath = outPath // fallback to relative path
	}
	fmt.Printf("✓ Generated: file://%s\n", absPath)
}

func printHelp() {
	help := `explain-html-gen - Generate interactive HTML explanations from structured content

Usage:
  explain-html-gen [flags] < input.json
  explain-html-gen --input input.json
  explain-html-gen --input input.json --output ./explanation.html

Flags:
  --input <file>     Path to JSON input file
  --stdin            Read JSON from stdin (default if no --input)
  --output <file>    Output HTML path (default: /tmp/YYYY-MM-DD-explanation-<slug>.html)
  --help             Show this help message

Input JSON schema:
  {
    "title": "string",                  // Required: page title
    "slug": "string",                   // Required: URL-safe slug for filename
    "summary": "string",                // Required: 1-2 sentence summary
    "date": "YYYY-MM-DD",               // Optional: defaults to today
    "author": "string",                 // Optional: author attribution
    "background": { ... },              // Required: background section
    "intuition": { ... },               // Required: intuition section
    "code": { ... },                    // Required: code section
    "quiz": [ ... ],                    // Required: exactly 5 quiz questions
    "diagrams": [ ... ]                 // Optional: additional diagrams
  }

For full schema details, see the README or package documentation.

Output:
  On success, prints the absolute file:// URL to the generated HTML.
  On failure, exits with status 1 and prints error details to stderr.

Example:
  cat content.json | explain-html-gen > /dev/null
  explain-html-gen --input content.json --output ./docs/explanation.html
`
	fmt.Println(help)
}
