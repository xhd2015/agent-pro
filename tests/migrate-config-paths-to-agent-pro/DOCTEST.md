# Migrate Config File Path Knowledge to agent-pro-dot-tool

## Version

0.0.2

## DSN (Domain Specific Notion)

**Config Path Functions** — exported functions returning `string` paths resolved
against `$HOME`. These functions live in agent-pro-dot-tool packages:
`agent/claude/config/`, `agent/codex/config/`, `agent/opencode/config/`,
`agent/opencode/skills/`, `agent/codex/skills/`.

All paths use `os.UserHomeDir()` — no hardcoded `~` or `/home/` strings.
The `HOME` environment variable is set during tests to a temp directory.

## Test Index

- **claude-config-paths/** — `SettingsPath()`, `JSONConfigPath()`, `GlobalSkillsDir()`
- **codex-config-paths/** — `DefaultConfigPath()`
- **opencode-config-paths/** — `GlobalUserConfigPath()`
- **opencode-skills-paths/** — `GlobalSkillDirs()`, `LocalSkillDirs(projectDir)`
- **codex-skills-paths/** — `GlobalSkillDirs()`, `LocalSkillDirs(projectDir)`

## How to Run

```sh
doctest vet ./tests/migrate-config-paths-to-agent-pro
doctest test ./tests/migrate-config-paths-to-agent-pro/...
```

```go
import (
	"fmt"
	"os"
	"testing"

	claudeconfig "github.com/xhd2015/agent-pro/agent/cli/claude/config"
	codexconfig "github.com/xhd2015/agent-pro/agent/codex/config"
	opencodeconfig "github.com/xhd2015/agent-pro/agent/opencode/config"
	opencodeSkills "github.com/xhd2015/agent-pro/agent/opencode/skills"
	codexSkills "github.com/xhd2015/agent-pro/agent/codex/skills"
)

type Request struct {
	Home       string
	TestCase   string
	ProjectDir string
}

type Response struct {
	Paths []string
	Error string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	origHome := os.Getenv("HOME")
	defer func() {
		if origHome == "" {
			os.Unsetenv("HOME")
		} else {
			os.Setenv("HOME", origHome)
		}
	}()

	if req.Home != "" {
		if err := os.Setenv("HOME", req.Home); err != nil {
			return nil, err
		}
	}

	switch req.TestCase {
	case "claude-all":
		return &Response{Paths: []string{
			claudeconfig.SettingsPath(),
			claudeconfig.JSONConfigPath(),
			claudeconfig.GlobalSkillsDir(),
		}}, nil
	case "codex-config":
		return &Response{Paths: []string{codexconfig.DefaultConfigPath()}}, nil
	case "opencode-global-user":
		return &Response{Paths: []string{opencodeconfig.GlobalUserConfigPath()}}, nil
	case "opencode-skills-global":
		dirs := opencodeSkills.GlobalSkillDirs()
		return &Response{Paths: dirs}, nil
	case "opencode-skills-local":
		dirs := opencodeSkills.LocalSkillDirs(req.ProjectDir)
		return &Response{Paths: dirs}, nil
	case "codex-skills-global":
		dirs := codexSkills.GlobalSkillDirs()
		return &Response{Paths: dirs}, nil
	case "codex-skills-local":
		dirs := codexSkills.LocalSkillDirs(req.ProjectDir)
		return &Response{Paths: dirs}, nil
	default:
		return nil, fmt.Errorf("unknown test case: %s", req.TestCase)
	}
}
```
