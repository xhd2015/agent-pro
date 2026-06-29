# Scenario

**Feature**: traceparse leaf `parse-line/pi/message-update`

```
trace JSONL -> adapter registry -> parsed event JSON
```

## Preconditions
- Mode and inputs are set for this leaf.

## Steps
1. Configure `Request` fields for this scenario.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RawLine = `{"type":"message_update","message":{"role":"assistant","content":[{"type":"text","text":"world"}]},"assistantMessageEvent":{"type":"text_delta","delta":"hello delta"}}`
	return nil
}
```
