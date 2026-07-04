# Scenario

**Feature**: `LLM_MOCK_RUN_FLAGS=--log-events <path>` enables log-events on `llm-mock-run-grok`

```
LLM_MOCK_RUN_FLAGS=--log-events session.jsonl
llm-mock-run-grok -> ParseRunFlags(env+argv) -> orchestrator -> mock --agent-events-file
fake grok -> curl mock twice -> >=2 message AgentEvents in session.jsonl
```

## Steps

1. Same config + events input as subcommand leaf.
2. `UseShortcut` true; `RunFlagsEnv` supplies `--log-events`; no CLI run flags on shortcut argv.
3. Fake grok curls mock twice.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.UseShortcut = true
	req.ConfigEnv = "file"
	req.FakeGrokCmd = fakeGrokCurlTwice
	req.OmitCLIRunFlags = true
	req.ConfigJSON = minimalMockConfigJSON(8080, `[
    {
      "request": {"role": "user", "content": "config-first-prompt", "index": -1},
      "response": {"content": "from-config", "finish_reason": "stop"}
    }
  ]`)
	req.EventsJSONL = `{"request":{"role":"user","content":"events-second-prompt","index":-1},"response":{"content":"from-events","finish_reason":"stop"}}` + "\n"
	req.LogEventsPath = filepath.Join(t.TempDir(), "env-shortcut-log-events.jsonl")
	req.RunFlagsEnv = "--log-events " + req.LogEventsPath
	return nil
}
```