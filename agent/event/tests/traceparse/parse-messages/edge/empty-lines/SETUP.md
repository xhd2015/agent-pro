# Scenario

**Feature**: traceparse leaf `parse-messages/edge/empty-lines`

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
	req.RawLines = nil
	req.CreatedAt = "2026-05-25T18:26:22.524536+08:00"
	return nil
}
```
