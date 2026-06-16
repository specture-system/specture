package new

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/specture-system/specture/internal/fs"
	gitpkg "github.com/specture-system/specture/internal/git"
	specpkg "github.com/specture-system/specture/internal/spec"
)

const (
	specFileName = "SPEC.md"
	planFileName = "PLAN.md"
)

// Options holds user choices for new file creation.
type Options struct {
	Title     string
	ParentRef string
	SpecRef   string
	Plan      bool
}

// NewCommandContext holds information needed to create a new spec or plan file.
type NewCommandContext struct {
	WorkDir      string
	SpecsDir     string
	ParentRef    string
	ParentPath   string
	Title        string
	Author       string
	Number       int
	FullRef      string
	Kind         string
	FileName     string
	RelativePath string
	FilePath     string
}

// NewContext creates a new NewCommandContext for spec or plan creation.
func NewContext(workDir string, opts Options) (*NewCommandContext, error) {
	title := strings.TrimSpace(opts.Title)
	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	if strings.TrimSpace(opts.ParentRef) != "" && strings.TrimSpace(opts.SpecRef) != "" {
		return nil, fmt.Errorf("--spec cannot be combined with --parent")
	}

	specsDir := filepath.Join(workDir, "specs")
	fileName := specFileName
	kind := "spec"
	if opts.Plan {
		fileName = planFileName
		kind = "plan"
	}

	author := getAuthor(workDir)
	slug := ToSlug(title)
	if slug == "" {
		return nil, fmt.Errorf("title must contain at least one letter or number")
	}

	parentRef := strings.TrimSpace(opts.ParentRef)
	specRef := strings.TrimSpace(opts.SpecRef)
	parentPath, parentFullRef, existingDir, number, fullRef, err := resolveTarget(specsDir, parentRef, specRef)
	if err != nil {
		return nil, err
	}

	dirName := fmt.Sprintf("%03d-%s", number, slug)
	baseDir := specsDir
	if parentPath != "" {
		baseDir = filepath.Dir(parentPath)
	}
	if existingDir != "" {
		baseDir = filepath.Dir(existingDir)
		dirName = filepath.Base(existingDir)
	}

	filePath := filepath.Join(baseDir, dirName, fileName)
	relativePath, err := filepath.Rel(specsDir, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute spec path: %w", err)
	}

	if fullRef == "" {
		fullRef = strconv.Itoa(number)
		if parentFullRef != "" {
			fullRef = parentFullRef + "." + fullRef
		}
	}

	return &NewCommandContext{
		WorkDir:      workDir,
		SpecsDir:     specsDir,
		ParentRef:    parentRef,
		ParentPath:   parentPath,
		Title:        title,
		Author:       author,
		Number:       number,
		FullRef:      fullRef,
		Kind:         kind,
		FileName:     fileName,
		RelativePath: relativePath,
		FilePath:     filePath,
	}, nil
}

// CreateFile creates the target SPEC.md or PLAN.md file.
func (c *NewCommandContext) CreateFile() error {
	content, err := RenderFile(c.Title, c.Author, c.FileName)
	if err != nil {
		return fmt.Errorf("failed to render %s: %w", c.FileName, err)
	}
	if err := fs.SafeWriteFile(c.FilePath, content); err != nil {
		return fmt.Errorf("failed to write %s: %w", c.FileName, err)
	}
	return nil
}

func resolveTarget(specsDir, parentRef, specRef string) (parentPath, parentFullRef, existingDir string, number int, fullRef string, err error) {
	if specRef != "" {
		return resolveExplicitTarget(specsDir, specRef)
	}

	if parentRef != "" {
		parentPath, err = specpkg.ResolvePath(specsDir, parentRef)
		if err != nil {
			return "", "", "", 0, "", fmt.Errorf("failed to resolve parent spec %q: %w", parentRef, err)
		}
		parentInfo, err := specpkg.Parse(parentPath)
		if err != nil {
			return "", "", "", 0, "", fmt.Errorf("failed to parse parent spec: %w", err)
		}
		parentFullRef = parentInfo.FullRef
	}

	number, err = FindNextSpecNumber(specsDir, parentPath)
	if err != nil {
		return "", "", "", 0, "", fmt.Errorf("failed to find next spec number: %w", err)
	}
	return parentPath, parentFullRef, "", number, "", nil
}

func resolveExplicitTarget(specsDir, specRef string) (parentPath, parentFullRef, existingDir string, number int, fullRef string, err error) {
	parts := strings.Split(specRef, ".")
	for _, part := range parts {
		if part == "" {
			return "", "", "", 0, "", fmt.Errorf("invalid spec reference: %s", specRef)
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return "", "", "", 0, "", fmt.Errorf("invalid spec reference: %s", specRef)
		}
	}

	last, _ := strconv.Atoi(parts[len(parts)-1])
	normalized := normalizeRefParts(parts)
	if existingPath, err := specpkg.ResolvePath(specsDir, normalized); err == nil {
		existingInfo, err := specpkg.Parse(existingPath)
		if err != nil {
			return "", "", "", 0, "", fmt.Errorf("failed to parse existing spec: %w", err)
		}
		return "", "", filepath.Dir(existingPath), existingInfo.Number, existingInfo.FullRef, nil
	}

	if len(parts) == 1 {
		return "", "", "", last, normalized, nil
	}

	parentRef := strings.Join(normalizedRefParts(parts[:len(parts)-1]), ".")
	parentPath, err = specpkg.ResolvePath(specsDir, parentRef)
	if err != nil {
		return "", "", "", 0, "", fmt.Errorf("failed to resolve parent spec %q: %w", parentRef, err)
	}
	parentInfo, err := specpkg.Parse(parentPath)
	if err != nil {
		return "", "", "", 0, "", fmt.Errorf("failed to parse parent spec: %w", err)
	}
	return parentPath, parentInfo.FullRef, "", last, normalized, nil
}

func normalizeRefParts(parts []string) string {
	return strings.Join(normalizedRefParts(parts), ".")
}

func normalizedRefParts(parts []string) []string {
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		n, _ := strconv.Atoi(part)
		normalized = append(normalized, strconv.Itoa(n))
	}
	return normalized
}

func getAuthor(workDir string) string {
	author, err := gitpkg.GetAuthor(workDir)
	if err == nil && strings.TrimSpace(author) != "" {
		return author
	}

	cmd := exec.Command("git", "config", "--global", "user.name")
	if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) != "" {
		return strings.TrimSpace(string(output))
	}

	return "Unknown"
}
