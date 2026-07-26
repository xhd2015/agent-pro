# Scenario

**Feature**: `--log-events` appends AgentEvent JSONL when mock serves prefix exchanges

```
fake grok -> curl mock twice (config + events prompts)
mock server -> log AgentEvent message per served response -> session.jsonl (>=2 lines)
```

## Steps

1. Config JSON with one exchange for `config-first-prompt` → response `from-config`.
2. Events input JSONL with one exchange for `events-second-prompt` → response `from-events`.
3. Fake grok curls mock twice via `GROK_MODELS_BASE_URL`.
4. `--log-events` points at a fresh `.jsonl` file in temp.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeGrokCmd = fakeGrokCurlTwice
	req.ConfigJSON = minimalMockConfigJSON(8080, `[
    {
      "request": {"role": "user", "content": "config-first-prompt", "index": -1},
      "response": {"content": "from-config", "finish_reason": "stop"}
    }
  ]`)
	req.EventsJSONL = `{"request":{"role":"user","content":"events-second-prompt","index":-1},"response":{"content":"from-events","finish_reason":"stop"}}` + "\n"
	req.LogEventsPath = filepath.Join(t.TempDir(), "session-log-events.jsonl")
	return nil
}
```