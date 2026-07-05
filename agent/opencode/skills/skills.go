package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	agentskills "github.com/xhd2015/agent-pro/agent/skills"
)

type SkillInfo = agentskills.SkillInfo

// SkillListResult holds skills grouped by scope (global vs project-local).
type SkillListResult struct {
	Global []SkillInfo
	Local  []SkillInfo
}

// GlobalSkillDirs returns the global opencode-compatible skill directories.
func GlobalSkillDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, ".config", "opencode", "skills") + string(os.PathSeparator),
		filepath.Join(home, ".claude", "skills") + string(os.PathSeparator),
		filepath.Join(home, ".agents", "skills") + string(os.PathSeparator),
	}
}

// LocalSkillDirs returns the project-local opencode-compatible skill directories.
func LocalSkillDirs(projectDir string) []string {
	return []string{
		filepath.Join(projectDir, ".opencode", "skills") + string(os.PathSeparator),
		filepath.Join(projectDir, ".claude", "skills") + string(os.PathSeparator),
		filepath.Join(projectDir, ".agents", "skills") + string(os.PathSeparator),
	}
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
	result := &SkillListResult{}

	for _, d := range GlobalSkillDirs() {
		skills, err := agentskills.ListDir(strings.TrimSuffix(d, string(os.PathSeparator)))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", d, err)
		}
		result.Global = append(result.Global, skills...)
	}

	for _, d := range LocalSkillDirs(projectDir) {
		skills, err := agentskills.ListDir(strings.TrimSuffix(d, string(os.PathSeparator)))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", d, err)
		}
		result.Local = append(result.Local, skills...)
	}

	return result, nil
}
