package validate

import (
	"bytes"
	"fmt"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	gmfrontmatter "go.abhg.dev/goldmark/frontmatter"
)

// ValidStatus contains the valid status values for a spec.
var ValidStatus = []string{"draft", "approved", "in-progress", "completed", "rejected"}

// Frontmatter represents the YAML frontmatter of a spec.
// This is distinct from internal/spec's frontmatter type because validation
// needs access to additional fields (e.g., Author) for validation rules.
type Frontmatter struct {
	Status string `yaml:"status"`
	Author string `yaml:"author"`
	Number *int   `yaml:"number"`
}

// Spec represents a parsed spec file for validation purposes.
//
// Note: The internal/spec package provides a higher-level SpecInfo type used
// for status display, task extraction, and spec discovery. This Spec type
// serves a different purpose: it retains the raw goldmark AST (Document) and
// validation-specific fields (HasTaskList, Source) needed by the validator.
// File discovery (FindAll, ResolvePath) is already delegated to internal/spec
// via cmd/validate.go.
type Spec struct {
	Path        string
	Frontmatter *Frontmatter
	Title       string
	HasTaskList bool
	Source      []byte
	Document    ast.Node
}

// ParseSpec parses a spec file and returns a Spec struct.
// For higher-level spec info (tasks, status inference), see internal/spec.Parse.
func ParseSpec(path string) (*Spec, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseSpecContent(path, content)
}

// ParseSpecContent parses spec content and returns a Spec struct.
// This parser retains the raw goldmark AST for validation rules.
// For higher-level spec info (tasks, status inference), see internal/spec.ParseContent.
func ParseSpecContent(path string, content []byte) (*Spec, error) {
	// Create goldmark parser with frontmatter and task list extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			&gmfrontmatter.Extender{},
			extension.TaskList,
		),
	)

	// Create a parser context to capture frontmatter
	ctx := parser.NewContext()

	// Parse the document
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader, parser.WithContext(ctx))

	spec := &Spec{
		Path:     path,
		Source:   content,
		Document: doc,
	}

	// Extract frontmatter from context
	spec.Frontmatter = extractFrontmatter(ctx)

	// Extract title (first H1 heading)
	spec.Title = extractTitle(doc, content)

	// Check for task list heading
	spec.HasTaskList = hasTaskList(doc, content)

	return spec, nil
}

// extractFrontmatter extracts the YAML frontmatter from the parser context
func extractFrontmatter(ctx parser.Context) *Frontmatter {
	fmData := gmfrontmatter.Get(ctx)
	if fmData == nil {
		return nil
	}

	var fm Frontmatter
	if err := fmData.Decode(&fm); err != nil {
		return nil
	}

	return &fm
}

// extractTitle extracts the first H1 heading from the document
func extractTitle(doc ast.Node, source []byte) string {
	var title string
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if heading, ok := n.(*ast.Heading); ok && heading.Level == 1 {
			// Get the text content of the heading
			var buf bytes.Buffer
			for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
				if textNode, ok := child.(*ast.Text); ok {
					buf.Write(textNode.Segment.Value(source))
				}
			}
			title = buf.String()
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return title
}

// hasTaskList checks if the document contains a "Task List" heading (H2)
func hasTaskList(doc ast.Node, source []byte) bool {
	found := false
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if heading, ok := n.(*ast.Heading); ok && heading.Level == 2 {
			// Get the text content of the heading
			var buf bytes.Buffer
			for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
				if textNode, ok := child.(*ast.Text); ok {
					buf.Write(textNode.Segment.Value(source))
				}
			}
			if buf.String() == "Task List" {
				found = true
				return ast.WalkStop, nil
			}
		}
		return ast.WalkContinue, nil
	})
	return found
}
