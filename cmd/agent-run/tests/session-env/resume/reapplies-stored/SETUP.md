# Scenario

**Feature**: resume without re-passing flags reapplies stored PATH, env, GROK_HOME

```
seed meta: prepend_paths=[tools], env=[FOO=bar], agent_runner_config_home=PATH
  -> resume --agent-runner-binary env-logger <id> "followup"
  -> child PATH prefix tools, FOO=bar, GROK_HOME=PATH
```

## Steps

1. Create tools + config-home dirs; seed bound+exited meta with stored fields.
2. Resume **without** `--prepend-path`, `-e`, or `--agent-runner-config-home`.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	prepareEnvLoggingResume(t, req)

	req.SessionID = "sess-env-reapply"
	req.PrependPathDir = absPath(t, filepath.Join(req.TempDir, "tools-stored"))
	if err := os.MkdirAll(req.PrependPathDir, 0755); err != nil {
		return err
	}
	req.AgentRunnerConfigHome = absPath(t, filepath.Join(req.TempDir, "grok-home-stored"))
	if err := os.MkdirAll(req.AgentRunnerConfigHome, 0755); err != nil {
		return err
	}

	req.SeedPrependPaths = []string{req.PrependPathDir}
	req.SeedEnv = []string{"FOO=bar"}
	req.SeedConfigHome = req.AgentRunnerConfigHome
	seedBoundExitedMeta(t, req)

	req.Prompt = "resume reapply stored"
	// Intentionally omit prepend-path / -e / agent-runner-config-home.
	req.Args = []string{
		"resume",
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.SessionID,
		req.Prompt,
	}
	return nil
}
```
