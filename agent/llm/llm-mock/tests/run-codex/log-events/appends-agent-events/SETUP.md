# Scenario

**Feature**: `--log-events` with `--mock-events-preset=think-tool-message` appends think and message AgentEvents

```
llm-mock run --mock-events-preset=think-tool-message --log-events f.jsonl codex
orchestrator -> mock genQueue [think, tool_call, message]
fake codex -> curl /v1/responses once -> log AgentEvents
```

## Steps

1. No config env (default empty `exchanges[]`).
2. `--mock-events-preset=think-tool-message` and `--log-events` for AgentEvent proof.
3. Fake codex curls mock once via base URL from `$CODEX_HOME/config.toml`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ConfigJSON = ""
	req.ConfigEnv = ""
	req.MockEventsPreset = "think-tool-message"
	req.LogEventsPath = filepath.Join(t.TempDir(), "events.jsonl")
	req.FakeCodexCmd = fakeCodexCurlResponsesOnce
	req.ExpectedExit = 0
	return nil
}
```