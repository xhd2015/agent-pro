# Scenario

**Feature**: `run` applies prepend/env/config-home to the TTY child and persists meta

```
agent-run run --agent-runner grok-tty \
  --prepend-path DIR -e KEY=VALUE --agent-runner-config-home PATH \
  --agent-runner-binary <env-logger> "prompt"
  -> child Env + sessions/<id>/meta.json
```

## Steps

1. Grouping prefixes common TTY run defaults (leaves append flags/prompt).
2. Leaves write env-logging fake runner and finalize args.
3. Assert probe file + meta.json (or error path for non-TTY / soft-missing).

```go
import (
	"path/filepath"
	"testing"
)

// prepareEnvLoggingRun writes the env-logger fake runner and records probe path.
func prepareEnvLoggingRun(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.TempDir, "fake-bin")
	req.EnvProbePath = filepath.Join(req.TempDir, "env-probe.log")
	req.RunnerScriptPath = writeEnvLoggingRunner(t, binDir, req.EnvProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath
	if req.Prompt == "" {
		req.Prompt = "session env probe"
	}
}

func Setup(t *testing.T, req *Request) error {
	// Default TTY run prefix; leaves append flags, binary, and prompt.
	req.Args = []string{"run", "--agent-runner", "grok-tty"}
	return nil
}
```
