# Scenario

**Feature**: config home seeds grok session discovery

```
temp config home + updates.jsonl
  -> --agent-runner-config-home PATH
  -> stderr grok-tty: grok session <uuid>
  -> stdout streams seeded assistant marker
```

## Steps

1. Seed fake grok session under temp config home with fixed UUID.
2. Run with `--agent-runner-config-home` and hold fake runner binary.

```go
import "testing"

const configHomeUUID = "b2222222-2222-4222-8222-222222222222"

func Setup(t *testing.T, req *Request) error {
	req.AgentRunnerConfigHome = filepath.Join(req.TempDir, "shared-grok-home")
	req.Prompt = "config home discovery"
	req.GrokSessionUUID = configHomeUUID
	req.GrokUpdatesPath = writeFakeGrokSessionDir(t, req.AgentRunnerConfigHome, req.TempDir, configHomeUUID, req.Prompt,
		acpAgentMessageChunk("CONFIG_HOME_STREAM_MARKER"),
	)

	binDir := filepath.Join(req.TempDir, "fake-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	req.RunnerScriptPath = writeHoldFakeRunner(t, binDir, "hold.sh", 2)
	req.AgentRunnerBinary = req.RunnerScriptPath
	req.Args = append(req.Args,
		"--agent-runner-config-home", req.AgentRunnerConfigHome,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.Prompt,
	)
	return nil
}
```