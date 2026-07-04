# Scenario

**Feature**: no `LLM_MOCK_EVENTS_FILE` — config-only exchanges suffice

```
mockconfig loader -> config exchanges only (no events file)
fake grok -> single curl -> from-config
```

## Steps

1. Do not set `EventsJSONL` (no `LLM_MOCK_EVENTS_FILE` env).
2. Fake grok performs one curl against config exchange.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.FakeGrokCmd = fakeGrokCurlOnce
	req.ConfigJSON = minimalMockConfigJSON(8080, `[
    {
      "request": {"role": "user", "content": "config-only-prompt", "index": -1},
      "response": {"content": "from-config", "finish_reason": "stop"}
    }
  ]`)
	return nil
}
```
