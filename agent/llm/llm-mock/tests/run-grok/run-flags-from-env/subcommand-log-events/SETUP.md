# Scenario

**Feature**: `LLM_MOCK_RUN_FLAGS=--log-events <path>` enables log-events on `llm-mock run grok` without CLI flag

```
LLM_MOCK_RUN_FLAGS=--log-events session.jsonl
llm-mock run grok -> orchestrator -> mock --agent-events-file session.jsonl
fake grok -> curl mock twice -> >=2 message AgentEvents in session.jsonl
```

## Steps

1. Config JSON with one exchange; events input JSONL with one more exchange.
2. Set `RunFlagsEnv` with `--log-events` and temp `.jsonl` path; omit CLI run flags.
3. Fake grok curls mock twice via `GROK_MODELS_BASE_URL`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
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
	req.LogEventsPath = filepath.Join(t.TempDir(), "env-subcommand-log-events.jsonl")
	req.RunFlagsEnv = "--log-events " + req.LogEventsPath
	return nil
}
```