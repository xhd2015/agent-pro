# Scenario

**Feature**: PTY child receives `GROK_HOME` only (not `AGENT_RUNNER_CONFIG_HOME`)

```
--agent-runner-config-home PATH
  -> child argv env GROK_HOME=PATH ...
  -> child process env dump lacks AGENT_RUNNER_CONFIG_HOME=
```

## Steps

1. Write env-logging fake runner.
2. Run with `--agent-runner-config-home` and `--agent-runner-binary`.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.AgentRunnerConfigHome = filepath.Join(req.TempDir, "child-env-home")
	if err := os.MkdirAll(req.AgentRunnerConfigHome, 0755); err != nil {
		return err
	}
	binDir := filepath.Join(req.TempDir, "fake-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	req.EnvProbePath = filepath.Join(req.TempDir, "env-probe.log")
	req.RunnerScriptPath = writeEnvLoggingRunner(t, binDir, req.EnvProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath
	req.Prompt = "env probe"
	req.Args = append(req.Args,
		"--agent-runner-config-home", req.AgentRunnerConfigHome,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.Prompt,
	)
	return nil
}
```