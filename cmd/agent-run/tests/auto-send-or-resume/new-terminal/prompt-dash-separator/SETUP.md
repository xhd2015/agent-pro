# Scenario

**Feature**: prompt starting with `-` is placed after `--` in the iTerm follow-up
command so it is not parsed as flags

```
no meta
  -> run --auto-send-or-resume --new-terminal --session-id ID -- -v explain
  -> exit 0; script follow-up contains `--` then prompt text (-v explain)
```

## Steps

1. Missing session (MODE=run) so ForceNew path is taken.
2. Pass a dash-leading multi-token prompt.
3. Capture iTerm script; assert `--` separator before prompt.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "nt-prompt-dash-d4"
	// Multi-token dash-leading prompt (joined with space as agent-run does).
	req.FollowupPrompt = "-v explain"
	req.WorkDir = req.TempDir
	req.Workspace = req.TempDir

	// Negative spawn probe.
	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe-launcher-must-not-write.log")
	binDir := filepath.Join(req.TempDir, "fake-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	req.RunnerScriptPath = writeArgvRecordingRunner(t, binDir, "record-argv.sh", req.ArgvProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath

	ensureItermScriptOutPath(req)
	_ = os.Remove(req.ItermScriptOut)

	// Pass prompt as remain after -- so -v is not parsed as a flag
	// (scenario: run … -- -v explain).
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--new-terminal",
		"--session-id", req.SessionID,
		"--agent-runner-binary", req.AgentRunnerBinary,
		"--",
		"-v", "explain",
	}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
