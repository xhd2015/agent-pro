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
	"github.com/xhd2015/doctest/session"
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

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	// Residual process Setenv(HOME): product path helpers (SettingsPath,
	// DefaultConfigPath, GlobalSkillDirs, …) call os.UserHomeDir() with no
	// home-parameter option. No small product API path yet across packages.
	// Restore after Run so other parallel leaves are less likely to observe
	// a leaked HOME (still a race under t.Parallel — product fix preferred).
	origHome, hadHome := os.LookupEnv("HOME")
	if req.Home != "" {
		if err := os.Setenv("HOME", req.Home); err != nil {
			return nil, err
		}
	}
	defer func() {
		if !hadHome {
			_ = os.Unsetenv("HOME")
		} else {
			_ = os.Setenv("HOME", origHome)
		}
	}()

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
