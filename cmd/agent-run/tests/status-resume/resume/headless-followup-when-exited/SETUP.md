# Scenario

**Feature**: resume headless with followup when bound+exited invokes run path with `--resume <id>`

```
seed finished bound+exited meta
  -> agent-run resume --agent-runner-binary <argv-recorder> <id> "followup"
  -> exit 0; argv probe contains --resume <runner_session_id>
```

## Steps

1. Seed bound exited meta (no live terminal).
2. Write argv-recording fake runner; do **not** set `AGENT_RUN_GROK_TTY_COMMAND`.
3. Run resume with `--agent-runner-binary` and followup prompt.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "test-resume-ok-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440500"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-resume-ok-1"
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
	req.FollowupPrompt = "resume followup please"
	// Ensure hook does not replace argv (would hide --resume).
	req.GrokTTYCommand = ""
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	req.Args = []string{
		"resume",
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.SessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
