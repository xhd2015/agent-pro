# Scenario

**Feature**: `resume --grok-session-id` resolves bound+exited meta and runs the
same success path as bare-id resume (provider `--resume` in argv)

```
seed finished bound+exited grok-tty + argv-recording fake runner
  -> agent-run resume --agent-runner-binary <recorder> --grok-session-id UUID "followup"
  -> exit 0; argv probe contains --resume <runner_session_id>
```

## Steps

1. Seed bound exited meta (no live terminal) with known UUID.
2. Write argv-recording fake runner; do **not** set `AGENT_RUN_GROK_TTY_COMMAND`.
3. Run resume with `--grok-session-id` (no positional session id) and followup.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Runner = "grok-tty"
	req.SessionID = "test-resume-gsid-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440910"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-resume-gsid-1"
	req.InitialPrompt = "prior turn"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	binDir := filepath.Join(req.TempDir, "fake-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe.log")
	req.RunnerScriptPath = writeArgvRecordingRunner(t, binDir, "record-argv.sh", req.ArgvProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath
	req.FollowupPrompt = "resume via grok-session-id"
	// Ensure hook does not replace argv (would hide --resume).
	req.GrokTTYCommand = ""
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	req.Args = []string{
		"resume",
		"--agent-runner-binary", req.AgentRunnerBinary,
		"--grok-session-id", req.RunnerSessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
