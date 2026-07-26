# Scenario

**Feature**: auto + `--new-terminal` + bound+exited (MODE=resume) opens iTerm2
ForceNew; parent does not spawn resume provider

```
seed finished bound+exited meta
  -> run --auto-send-or-resume --new-terminal --session-id ID … "followup"
  -> exit 0; KOOL_ITERM2_SCRIPT_OUT written (create window)
  -> no in-process --resume argv probe
```

## Steps

1. Seed bound exited meta (no live terminal).
2. Install argv-recording runner as negative probe.
3. Enable iTerm script capture; run auto + new-terminal + followup.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "nt-resume-d2"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440d22"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-nt-resume-d2"
	req.InitialPrompt = "prior turn nt-d2"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	installArgvRunner(t, req)
	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe-launcher-must-not-write.log")
	binDir := filepath.Join(req.TempDir, "fake-bin")
	req.RunnerScriptPath = writeArgvRecordingRunner(t, binDir, "record-argv.sh", req.ArgvProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath

	req.FollowupPrompt = "resume followup in iterm"
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
	req.Mode = "read-probes"
	return nil
}
```
