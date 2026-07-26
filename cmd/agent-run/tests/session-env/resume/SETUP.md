# Scenario

**Feature**: resume reapplies stored prepend/env/config-home; optional flags append

```
seed bound+exited meta with prepend_paths / env / agent_runner_config_home
  -> agent-run resume [--prepend-path|--env] <id> "followup"
  -> child Env from stored (+ append); meta updated
```

## Steps

1. Grouping provides helpers to seed bound+exited meta and env-logging binary.
2. Leaves set seed fields + resume args (with or without extra flags).
3. Assert probe + meta after resume.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

// prepareEnvLoggingResume installs env-logger as --agent-runner-binary for resume.
func prepareEnvLoggingResume(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.TempDir, "fake-bin")
	req.EnvProbePath = filepath.Join(req.TempDir, "env-probe-resume.log")
	req.RunnerScriptPath = writeEnvLoggingRunner(t, binDir, req.EnvProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath
	if req.Prompt == "" {
		req.Prompt = "resume followup"
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Resume grouping defaults; leaves seed meta and finalize Args.
	if req.SessionID == "" {
		req.SessionID = "sess-env-resume"
	}
	if req.Runner == "" {
		req.Runner = defaultRunner
	}
	return nil
}
```
