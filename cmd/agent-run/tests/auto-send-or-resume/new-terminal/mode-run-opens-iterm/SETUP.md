# Scenario

**Feature**: auto + `--new-terminal` + missing session (MODE=run) opens iTerm2
ForceNew and does not spawn the provider in the launcher process

```
no meta for id
  -> run --auto-send-or-resume --new-terminal --session-id NEW … "prompt"
  -> exit 0
  -> KOOL_ITERM2_SCRIPT_OUT: create window; follow-up has agent-run, auto, session-id,
     prompt after `--`; no --new-terminal
  -> no in-process provider argv probe
```

## Steps

1. Do **not** seed meta (MODE=run).
2. Install argv-recording fake runner path only as a **negative** probe (must stay empty).
3. Enable iTerm script capture; pass through `--agent-runner-binary` so reconstruct keeps it.
4. Run auto + new-terminal + session-id + prompt.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "nt-run-missing-d1"
	req.FollowupPrompt = "create in iterm please"
	req.WorkDir = req.TempDir
	req.Workspace = req.TempDir

	// Negative probe: launcher must not spawn provider in-process.
	installArgvRunner(t, req)
	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe-launcher-must-not-write.log")
	// Re-bind runner to write to the negative probe path if anything spawns.
	binDir := filepath.Join(req.TempDir, "fake-bin")
	req.RunnerScriptPath = writeArgvRecordingRunner(t, binDir, "record-argv.sh", req.ArgvProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath

	ensureItermScriptOutPath(req)
	_ = os.Remove(req.ItermScriptOut)

	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--new-terminal",
		"--session-id", req.SessionID,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
