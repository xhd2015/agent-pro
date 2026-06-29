# Scenario

**Feature**: traceparse leaf `parse-line/opencode/tool-bash`

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
	req.RawLine = `{"type":"tool_use","part":{"type":"tool","tool":"bash","state":{"status":"completed","input":{"command":"go test ./..."}}}}`
	return nil
}
```
