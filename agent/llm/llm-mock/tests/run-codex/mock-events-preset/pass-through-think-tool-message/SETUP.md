# Scenario

**Feature**: orchestrator passes `--mock-events-preset=think-tool-message` to mock; fake codex curls once

```
llm-mock run --mock-events-preset=think-tool-message --log-events f.jsonl codex
orchestrator -> mock genQueue [think, tool_call, message]
fake codex curl -> think, tool_call, message AgentEvents in log-events
```

## Steps

1. No config env (default empty `exchanges[]`).
2. `--mock-events-preset=think-tool-message` and `--log-events` for AgentEvent proof.
3. Fake codex curls mock once via base URL from `$CODEX_HOME/config.toml`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = ""
	req.ConfigEnv = ""
	req.MockEventsPreset = "think-tool-message"
	req.LogEventsPath = filepath.Join(t.TempDir(), "events.jsonl")
	req.FakeCodexCmd = fakeCodexCurlResponsesOnce
	req.ExpectedExit = 0
	return nil
}
```