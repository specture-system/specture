package setup

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/specture-system/specture/internal/templates"
)

// skillFiles defines the embedded skill files to install, rooted under "files/skills".
// The source paths are relative to the embed.FS root; the destination paths are
// relative to ".skills/" in the working directory.
var skillFiles = []struct {
	src string // path within embed.FS (under files/skills/)
	dst string // path relative to .skills/ in workDir
}{
	{"files/skills/specture/SKILL.md", "specture/SKILL.md"},
	{"files/skills/specture/references/spec-format.md", "specture/references/spec-format.md"},
}

// InstallSkill copies the embedded Specture skill files into the .skills/
// directory under workDir. Existing files are overwritten on re-run.
// When dryRun is true, no files are written — only a summary is printed.
func InstallSkill(workDir string, dryRun bool) error {
	skillFS := templates.GetSkillFiles()

	for _, sf := range skillFiles {
		dstPath := filepath.Join(workDir, ".skills", sf.dst)

		if dryRun {
			fmt.Printf("[dry-run] Would install skill file: %s\n", dstPath)
			continue
		}

		content, err := fs.ReadFile(skillFS, sf.src)
		if err != nil {
			return fmt.Errorf("failed to read embedded skill file %s: %w", sf.src, err)
		}

		// Ensure parent directory exists
		dir := filepath.Dir(dstPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Write (overwrite) the file
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write skill file %s: %w", dstPath, err)
		}
	}

	return nil
}
