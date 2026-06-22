// Package spec provides shared spec parsing, discovery, and querying.
package spec

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	gmfrontmatter "go.abhg.dev/goldmark/frontmatter"
)

const (
	specFilename = "SPEC.md"
	planFilename = "PLAN.md"
)

// SpecInfo represents a parsed spec file with all extracted metadata.
type SpecInfo struct {
	Path    string
	Name    string
	Number  int
	FullRef string
	Status  string
}

// frontmatter represents the YAML frontmatter of a spec.
type frontmatter struct {
	Status string `yaml:"status"`
}

// Parse reads and parses a spec file, returning a fully populated SpecInfo.
func Parse(path string) (*SpecInfo, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseContent(path, content)
}

// ParseContent parses spec content and returns a fully populated SpecInfo.
func ParseContent(path string, content []byte) (*SpecInfo, error) {
	info := &SpecInfo{
		Path: path,
	}

	// Parse with goldmark for frontmatter and title
	md := goldmark.New(
		goldmark.WithExtensions(
			&gmfrontmatter.Extender{},
		),
	)
	ctx := parser.NewContext()
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader, parser.WithContext(ctx))

	// Extract frontmatter
	var fmStatus string
	fmData := gmfrontmatter.Get(ctx)
	if fmData != nil {
		var fm frontmatter
		if err := fmData.Decode(&fm); err == nil {
			fmStatus = fm.Status
		}
	}

	number := extractNumberFromSpecPath(path)
	info.Number = number
	var err error
	info.FullRef, err = resolveFullRef(path, number)
	if err != nil {
		return nil, err
	}

	// Extract title (first H1 heading)
	info.Name = extractTitle(doc, content)

	// Status comes from frontmatter only.
	info.Status = inferStatus(fmStatus)

	return info, nil
}

// ParseAll finds and parses all specs in the given directory, sorted by ascending number.
func ParseAll(specsDir string) ([]*SpecInfo, error) {
	paths, err := FindAll(specsDir)
	if err != nil {
		return nil, err
	}

	var specs []*SpecInfo
	for _, p := range paths {
		info, err := Parse(p)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", p, err)
		}
		specs = append(specs, info)
	}

	sort.Slice(specs, func(i, j int) bool {
		return specs[i].Number < specs[j].Number
	})

	return specs, nil
}

// FindCurrent returns the first spec with status "in-progress", sorted by ascending number.
// Returns nil if no in-progress spec is found.
func FindCurrent(specs []*SpecInfo) *SpecInfo {
	// Make a sorted copy so we don't mutate the input
	sorted := make([]*SpecInfo, len(specs))
	copy(sorted, specs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Number < sorted[j].Number
	})

	for _, s := range sorted {
		if s.Status == "in-progress" {
			return s
		}
	}
	return nil
}

func resolveFullRef(path string, number int) (string, error) {
	if number < 0 {
		return "", nil
	}

	parentSpecPath, err := findParentSpecPath(path)
	if err != nil {
		return "", err
	}
	if parentSpecPath != "" {
		parentInfo, err := Parse(parentSpecPath)
		if err != nil {
			return "", err
		}
		if parentInfo.FullRef != "" {
			return parentInfo.FullRef + "." + strconv.Itoa(number), nil
		}
		return strconv.Itoa(number), nil
	}

	return strconv.Itoa(number), nil
}

// extractNumberFromSpecPath extracts the local spec number from the path.
// For nested specs it reads the parent directory name, and for legacy flat
// specs it falls back to the filename.
func extractNumberFromSpecPath(path string) int {
	target := filepath.Base(path)
	if IsSpecFilePath(path) {
		target = filepath.Base(filepath.Dir(path))
	}

	re := regexp.MustCompile(`^(\d+)`)
	matches := re.FindStringSubmatch(target)
	if len(matches) < 2 {
		return -1
	}
	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return -1
	}
	return n
}

// extractTitle extracts the first H1 heading text from a goldmark document.
func extractTitle(doc ast.Node, source []byte) string {
	var title string
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if heading, ok := n.(*ast.Heading); ok && heading.Level == 1 {
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

// inferStatus determines the spec status from frontmatter only.
func inferStatus(fmStatus string) string {
	if fmStatus != "" {
		return fmStatus
	}

	return "draft"
}

// FindAll finds all nested spec files in the given specs directory.
func FindAll(specsDir string) ([]string, error) {
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("specs directory not found: %s", specsDir)
	}

	var paths []string
	if err := collectSpecPaths(specsDir, &paths); err != nil {
		return nil, err
	}

	sort.Strings(paths)
	return paths, nil
}

// ResolvePath resolves a spec reference or spec file path to a file path.
// Accepts:
//   - Full path to a SPEC.md or PLAN.md file
//   - Top-level references with or without leading zeros: 0, 00, 000
//   - Hierarchical references: 1.4.3
//
// Lookups are performed against the parsed full reference derived from the
// directory tree.
func ResolvePath(specsDir, arg string) (string, error) {
	// If it's already a path that exists, use it
	if _, err := os.Stat(arg); err == nil {
		if !IsSpecFilePath(arg) {
			return "", fmt.Errorf("spec paths must point to a SPEC.md or PLAN.md file: %s", arg)
		}
		return arg, nil
	}

	fullRef, err := normalizeSpecRef(arg)
	if err != nil {
		return "", err
	}

	paths, err := FindAll(specsDir)
	if err != nil {
		return "", err
	}

	for _, p := range paths {
		info, err := Parse(p)
		if err != nil {
			continue
		}
		if info.FullRef == fullRef {
			return p, nil
		}
	}

	return "", fmt.Errorf("spec not found: %s", arg)
}

// FindSpecsInScope returns parsed specs that belong directly under the requested scope.
// With no parent path, it returns top-level specs under specsDir.
// With a parent path, it returns only immediate child specs of that parent.
func FindSpecsInScope(specsDir, parentPath string) ([]*SpecInfo, error) {
	return FindSpecsInScopeDepth(specsDir, parentPath, 1)
}

// FindSpecsInScopeDepth returns parsed specs within the given depth of the scope root.
// depth 1 returns immediate children only (same as FindSpecsInScope).
// depth <= 0 is treated as unlimited.
func FindSpecsInScopeDepth(specsDir, parentPath string, depth int) ([]*SpecInfo, error) {
	if parentPath != "" && !IsSpecFilePath(parentPath) {
		return nil, fmt.Errorf("parent spec must be a SPEC.md or PLAN.md file: %s", parentPath)
	}

	paths, err := FindAll(specsDir)
	if err != nil {
		return nil, err
	}

	// Filter paths by scope membership and depth. d < 1 means the spec is
	// either the scope root itself or not a descendant at all.
	var scopedPaths []string
	for _, path := range paths {
		d := specDepth(path, specsDir, parentPath)
		if d < 1 {
			continue
		}
		if depth > 0 && d > depth {
			continue
		}
		scopedPaths = append(scopedPaths, path)
	}

	var specs []*SpecInfo
	for _, path := range scopedPaths {
		info, err := Parse(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}
		specs = append(specs, info)
	}

	// Sort by FullRef, not Number. For nested specs, the local Number field
	// (from the directory prefix like "001-") doesn't reflect hierarchical
	// position — "4.1" has Number=1 but FullRef="4.1" and should sort
	// between "4" and "5", not next to the top-level "1".
	sort.Slice(specs, func(i, j int) bool {
		return compareFullRefs(specs[i].FullRef, specs[j].FullRef)
	})

	return specs, nil
}

// compareFullRefs compares two dotted numeric refs segment-by-segment.
// Needed because FullRef is a string like "4.1" or "10", and naive string
// comparison would place "10" before "4". Each segment is compared as an
// integer, so "4.1" < "4.2", "4" < "4.1", "10" > "4".
func compareFullRefs(a, b string) bool {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		aNum, _ := strconv.Atoi(aParts[i])
		bNum, _ := strconv.Atoi(bParts[i])
		if aNum != bNum {
			return aNum < bNum
		}
	}
	return len(aParts) < len(bParts)
}

// specDepth computes how many directory levels a spec is from the scope root.
// Uses filepath.Rel to count directory separators between root and the spec's
// parent directory. rel == "." means root == specDir (the scope root itself,
// depth 0). rel starting with ".." means the spec is outside the scope tree.
// Returns -1 for non-descendants.
func specDepth(specPath, specsDir, parentPath string) int {
	if !IsSpecFilePath(specPath) {
		return -1
	}

	specDir := filepath.Dir(specPath)

	var rootDir string
	if parentPath != "" {
		rootDir = filepath.Dir(parentPath)
	} else {
		rootDir = specsDir
	}

	rel, err := filepath.Rel(rootDir, specDir)
	if err != nil {
		return -1
	}

	if rel == "." {
		return 0
	}

	parts := strings.Split(rel, string(filepath.Separator))
	if parts[0] == ".." {
		return -1
	}

	return len(parts)
}

// IsTopLevelSpecPath reports whether a spec path belongs directly under specsDir.
func IsTopLevelSpecPath(specPath, specsDir string) bool {
	return IsSpecFilePath(specPath) &&
		filepath.Dir(filepath.Dir(specPath)) == specsDir
}

// IsImmediateChildSpecPath reports whether specPath is a direct child of parentPath.
// This is the path shape used for nested specs created under a parent spec directory.
func IsImmediateChildSpecPath(specPath, parentPath string) bool {
	return IsSpecFilePath(specPath) &&
		filepath.Dir(filepath.Dir(specPath)) == filepath.Dir(parentPath)
}

// collectSpecPaths walks the specs tree and records every discoverable spec file.
func collectSpecPaths(dir string, paths *[]string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read specs directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			if err := collectSpecPaths(path, paths); err != nil {
				return err
			}
			continue
		}

		if filepath.Base(path) == specFilename {
			*paths = append(*paths, path)
			continue
		}
		if filepath.Base(path) == planFilename {
			// Only use PLAN.md when a SPEC.md file is not present in the same directory.
			hasSpec := false
			for _, sibling := range entries {
				if sibling.IsDir() {
					continue
				}
				if sibling.Name() == specFilename {
					hasSpec = true
					break
				}
			}
			if !hasSpec {
				*paths = append(*paths, path)
			}
		}
	}

	return nil
}

// IsSpecFilePath reports whether path points to a supported spec file name.
func IsSpecFilePath(path string) bool {
	base := filepath.Base(path)
	return base == specFilename || base == planFilename
}

// findParentSpecPath returns the nearest supported spec file in the parent directory.
// If no parent spec exists, it returns an empty path.
func findParentSpecPath(path string) (string, error) {
	parentDir := filepath.Dir(filepath.Dir(path))
	if parentDir == "." || parentDir == string(filepath.Separator) {
		return "", nil
	}

	specPath := filepath.Join(parentDir, specFilename)
	if _, err := os.Stat(specPath); err == nil {
		return specPath, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to inspect parent spec: %w", err)
	}

	planPath := filepath.Join(parentDir, planFilename)
	if _, err := os.Stat(planPath); err == nil {
		return planPath, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to inspect parent spec: %w", err)
	}

	return "", nil
}

// normalizeSpecRef canonicalizes a user-provided reference so lookup can match
// against parsed FullRef values. It trims whitespace and removes leading zeros
// from each segment, so values like 001.002 compare as 1.2.
func normalizeSpecRef(arg string) (string, error) {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" {
		return "", fmt.Errorf("invalid spec reference: %s (expected a reference like 0, 00, 000, or 1.4.3)", arg)
	}

	parts := strings.Split(trimmed, ".")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return "", fmt.Errorf("invalid spec reference: %s (expected a reference like 0, 00, 000, or 1.4.3)", arg)
		}

		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return "", fmt.Errorf("invalid spec reference: %s (expected a reference like 0, 00, 000, or 1.4.3)", arg)
		}
		normalized = append(normalized, strconv.Itoa(number))
	}

	return strings.Join(normalized, "."), nil
}
