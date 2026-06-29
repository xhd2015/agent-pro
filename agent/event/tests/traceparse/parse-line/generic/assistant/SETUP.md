# Scenario

**Feature**: traceparse leaf `parse-line/generic/assistant`

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
	req.RawLine = `{"type":"assistant","message":{"content":[{"type":"text","text":"generic reply"}]}}`
	return nil
}
```
