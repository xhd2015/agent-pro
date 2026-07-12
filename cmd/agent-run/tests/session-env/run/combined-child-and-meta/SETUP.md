# Scenario

**Feature**: combined `--prepend-path` + `-e` + `--agent-runner-config-home` on run

```
run --prepend-path tools -e FOO=bar --agent-runner-config-home PATH \
  --agent-runner-binary env-logger "prompt"
  -> child PATH starts with abs(tools)
  -> child FOO=bar, GROK_HOME=PATH (no AGENT_RUNNER_CONFIG_HOME= required)
  -> meta.json has prepend_paths, env, agent_runner_config_home (abs)
```

## Steps

1. Create tools dir and config-home dir under temp.
2. Write env-logging fake runner.
3. Run with all three flag families and a stable `--session-id`.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	prepareEnvLoggingRun(t, req)

	req.PrependPathDir = filepath.Join(req.TempDir, "tools")
	if err := os.MkdirAll(req.PrependPathDir, 0755); err != nil {
		return err
	}
	req.PrependPathDir = absPath(t, req.PrependPathDir)

	req.AgentRunnerConfigHome = filepath.Join(req.TempDir, "grok-home")
	if err := os.MkdirAll(req.AgentRunnerConfigHome, 0755); err != nil {
		return err
	}
	req.AgentRunnerConfigHome = absPath(t, req.AgentRunnerConfigHome)

	req.SessionID = "sess-env-combined"
	req.Prompt = "combined env probe"
	req.Args = append(req.Args,
		"--session-id", req.SessionID,
		"--prepend-path", req.PrependPathDir,
		"-e", "FOO=bar",
		"--agent-runner-config-home", req.AgentRunnerConfigHome,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.Prompt,
	)
	return nil
}
```
