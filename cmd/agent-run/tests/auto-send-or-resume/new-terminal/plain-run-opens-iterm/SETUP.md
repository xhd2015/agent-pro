# Scenario

**Feature**: `--new-terminal` without `--auto-send-or-resume` ForceNew-opens iTerm
and re-invokes `run` with the flag stripped (new session in a new window)

```
agent-run run --new-terminal "hi"
  -> exit 0
  -> KOOL_ITERM2_SCRIPT_OUT: create window; follow-up has run + prompt after `--`;
     no --new-terminal
  -> no in-process provider spawn
```

## Steps

1. Enable iTerm script capture.
2. Negative argv probe.
3. Run `run --new-terminal` with a prompt and no auto flag.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FollowupPrompt = "hi"
	req.WorkDir = req.TempDir
	req.Workspace = req.TempDir

	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe-launcher-must-not-write.log")
	installArgvRunner(t, req)

	ensureItermScriptOutPath(req)
	_ = os.Remove(req.ItermScriptOut)

	req.Args = []string{
		"run",
		"--new-terminal",
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
