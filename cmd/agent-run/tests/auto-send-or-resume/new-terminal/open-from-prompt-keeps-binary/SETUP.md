# Scenario

**Feature**: `--open --new-terminal --session-id-from-prompt` keeps binary,
config-home, and `--env` on the iTerm child (new interactive session)

```
run --open --new-terminal --session-id-from-prompt \
  --agent-runner=codex-tty --agent-runner-binary BIN \
  --agent-runner-config-home HOME --env LLM_MOCK_MCP=slow_01=1s-10s \
  -- "wait then say done"
  -> exit 0
  -> follow-up has --open, --session-id-from-prompt, binary, config-home, MCP env
  -> no --new-terminal; launcher does not spawn provider
```

## Steps

1. Fake runner as negative spawn probe.
2. iTerm script capture.
3. Args match the interactive mock-mcp recipe (minus scene mktemp).

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FollowupPrompt = "wait then say done"
	req.WorkDir = req.TempDir
	req.Workspace = req.TempDir
	req.DirOverride = req.TempDir

	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe-launcher-must-not-write.log")
	installArgvRunner(t, req)

	configHome := filepath.Join(req.TempDir, "codex-home")
	if err := os.MkdirAll(configHome, 0755); err != nil {
		return err
	}

	ensureItermScriptOutPath(req)
	_ = os.Remove(req.ItermScriptOut)

	req.Args = []string{
		"run",
		"--open",
		"--new-terminal",
		"--color",
		"--agent-runner=codex-tty",
		"--agent-runner-binary", req.AgentRunnerBinary,
		"--agent-runner-config-home", configHome,
		"--session-id-from-prompt",
		"--dir", req.DirOverride,
		"--env", "LLM_MOCK_MCP=slow_01=1s-10s",
		"--",
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
