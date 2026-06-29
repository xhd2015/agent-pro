# Scenario

**Feature**: traceparse leaf `parse-messages/cursor/merge-lifecycle`

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
	req.RawLines = []string{
		`{"type":"tool_call","subtype":"started","call_id":"cursor_1","tool_call":{"shellToolCall":{"args":{"command":"go test ./..."}}}}`,
		`{"type":"tool_call","subtype":"completed","call_id":"cursor_1","tool_call":{"shellToolCall":{"result":{"exit_code":0,"output":"ok"}}}}`,
	}
	req.CreatedAt = "2026-05-25T18:26:22.524536+08:00"
	return nil
}
```
