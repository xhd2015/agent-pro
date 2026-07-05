# Scenario

**Feature**: `--agent-runner-binary` fake script receives grok-tty default argv

```
agent-run run --agent-runner-binary <script> "argv probe"
  -> PTY runs script (not AGENT_RUN_GROK_TTY_COMMAND)
  -> stderr ARGV_RECORD includes --always-approve
```

## Steps

1. Write argv-recording fake runner script.
2. Run with `--agent-runner-binary` pointing at the script; hook env unset.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	binDir := filepath.Join(req.TempDir, "fake-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe.log")
	req.RunnerScriptPath = writeArgvRecordingRunner(t, binDir, "record-argv.sh", req.ArgvProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath
	req.Prompt = "argv probe"
	req.Args = append(req.Args, "--agent-runner-binary", req.AgentRunnerBinary, req.Prompt)
	return nil
}
```