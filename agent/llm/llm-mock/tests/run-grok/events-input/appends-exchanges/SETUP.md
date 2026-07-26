# Scenario

**Feature**: events file appends second exchange after config exchanges

```
mockconfig loader -> config exchange[0] + events JSONL exchange[1]
fake grok -> curl mock twice -> R2 content from events file
```

## Steps

1. Config JSON with one exchange matching `config-first-prompt` → `from-config`.
2. Events JSONL with one exchange matching `events-second-prompt` → `from-events`.
3. Fake grok curls twice via `GROK_MODELS_BASE_URL`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FakeGrokCmd = fakeGrokCurlTwice
	req.ConfigJSON = minimalMockConfigJSON(8080, `[
    {
      "request": {"role": "user", "content": "config-first-prompt", "index": -1},
      "response": {"content": "from-config", "finish_reason": "stop"}
    }
  ]`)
	req.EventsJSONL = `{"request":{"role":"user","content":"events-second-prompt","index":-1},"response":{"content":"from-events","finish_reason":"stop"}}` + "\n"
	return nil
}
```
