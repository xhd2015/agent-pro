package skills

import (
	"fmt"
	"os"
	"path/filepath"

	agentskills "github.com/xhd2015/agent-pro/agent/skills"
)

type SkillInfo = agentskills.SkillInfo

// SkillListResult holds skills grouped by scope (global vs project-local).
type SkillListResult struct {
	Global []SkillInfo
	Local  []SkillInfo
}

// List returns skills from all opencode-compatible skill directories.
//
// Global directories:
//   - ~/.config/opencode/skills/
//   - ~/.claude/skills/
//   - ~/.agents/skills/
//
// Project-local directories (under projectDir):
//   - .opencode/skills/
//   - .claude/skills/
//   - .agents/skills/
func List(projectDir string) (*SkillListResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("find home dir: %w", err)
	}

	result := &SkillListResult{}

	// Global directories
	globalDirs := []string{
		filepath.Join(home, ".config", "opencode", "skills"),
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".agents", "skills"),
	}
	for _, d := range globalDirs {
		skills, err := agentskills.ListDir(d)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", d, err)
		}
		result.Global = append(result.Global, skills...)
	}

	// Project-local directories
	localDirs := []string{
		filepath.Join(projectDir, ".opencode", "skills"),
		filepath.Join(projectDir, ".claude", "skills"),
		filepath.Join(projectDir, ".agents", "skills"),
	}
	for _, d := range localDirs {
		skills, err := agentskills.ListDir(d)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", d, err)
		}
		result.Local = append(result.Local, skills...)
	}

	return result, nil
}
