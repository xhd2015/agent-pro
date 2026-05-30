package skills

import (
	"fmt"
	"os"
	"path/filepath"

	agentskills "github.com/xhd2015/agent-pro/agent/skills"
)

type SkillInfo = agentskills.SkillInfo

// SkillListResult holds skills grouped by scope.
type SkillListResult struct {
	Global []SkillInfo // from ~/.codex/skills/ and ~/.agents/skills/
	Local  []SkillInfo // from .agents/skills/ (project-local)
}

// List returns skills from Codex and agent-skills-standard directories.
//
// Codex reads skills from:
//   - ~/.codex/skills/           (Codex legacy directory)
//   - ~/.agents/skills/          (open agent skills standard, user-level)
//   - .agents/skills/            (open agent skills standard, repo-level)
func List(projectDir string) (*SkillListResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("find home dir: %w", err)
	}

	result := &SkillListResult{}

	// Codex-native global directory
	codexDir := filepath.Join(home, ".codex", "skills")
	skills, err := agentskills.ListDir(codexDir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", codexDir, err)
	}
	result.Global = append(result.Global, skills...)

	// Open agent skills standard: global (~/.agents/skills/)
	agentsGlobalDir := filepath.Join(home, ".agents", "skills")
	skills, err = agentskills.ListDir(agentsGlobalDir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", agentsGlobalDir, err)
	}
	result.Global = append(result.Global, skills...)

	// Open agent skills standard: project-local (.agents/skills/)
	agentsLocalDir := filepath.Join(projectDir, ".agents", "skills")
	skills, err = agentskills.ListDir(agentsLocalDir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", agentsLocalDir, err)
	}
	result.Local = append(result.Local, skills...)

	return result, nil
}
