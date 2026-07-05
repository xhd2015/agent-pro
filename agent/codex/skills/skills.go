package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	agentskills "github.com/xhd2015/agent-pro/agent/skills"
)

type SkillInfo = agentskills.SkillInfo

// SkillListResult holds skills grouped by scope.
type SkillListResult struct {
	Global []SkillInfo // from ~/.codex/skills/ and ~/.agents/skills/
	Local  []SkillInfo // from .agents/skills/ (project-local)
}

// GlobalSkillDirs returns the global Codex-compatible skill directories.
func GlobalSkillDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, ".codex", "skills") + string(os.PathSeparator),
		filepath.Join(home, ".agents", "skills") + string(os.PathSeparator),
	}
}

// LocalSkillDirs returns the project-local Codex-compatible skill directories.
func LocalSkillDirs(projectDir string) []string {
	return []string{
		filepath.Join(projectDir, ".agents", "skills") + string(os.PathSeparator),
	}
}

// List returns skills from Codex and agent-skills-standard directories.
//
// Codex reads skills from:
//   - ~/.codex/skills/           (Codex legacy directory)
//   - ~/.agents/skills/          (open agent skills standard, user-level)
//   - .agents/skills/            (open agent skills standard, repo-level)
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
