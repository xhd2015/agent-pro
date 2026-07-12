# Scenario

**Bug**: headless `resume <id> "followup"` with zombie registry same id must
not fail `session id already in use` (reclaim + run with `--resume`)

```
seed zombie: session_id == terminal_session_id
  -> agent-run resume --agent-runner-binary <argv-recorder> <id> "followup"
  -> reclaim zombie -> reserve same id -> argv includes --resume <runner_session_id>
  -> NOT: already in use
```

## Steps

1. Start detached sleep as zombie serve PID.
2. Seed bound+exited zombie fixture with SessionID == TerminalSessionID.
3. Write argv-recording runner (no `AGENT_RUN_GROK_TTY_COMMAND` hook).
4. Run headless resume with followup.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "test-resume-zombie-s1"
	req.TerminalSessionID = "test-resume-zombie-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440702"
	req.InitialPrompt = "prior zombie turn"
	req.RegistryPID = startDetachedSleepPID(t)
	seedZombieServeAfterExit(t, req)

	binDir := filepath.Join(req.TempDir, "fake-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe.log")
	req.RunnerScriptPath = writeArgvRecordingRunner(t, binDir, "record-argv.sh", req.ArgvProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath
	req.FollowupPrompt = "resume after zombie reclaim"
	req.GrokTTYCommand = ""
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	req.Args = []string{
		"resume",
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.SessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	req.Mode = "read-meta"
	return nil
}
```
