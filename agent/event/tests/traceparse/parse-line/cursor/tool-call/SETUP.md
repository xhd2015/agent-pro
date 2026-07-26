# Scenario

**Feature**: traceparse leaf `parse-line/cursor/tool-call`

```
trace JSONL -> adapter registry -> parsed event JSON
```

## Preconditions
- Mode and inputs are set for this leaf.

## Steps
1. Configure `Request` fields for this scenario.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RawLine = `{"type":"tool_call","subtype":"completed","call_id":"cursor_1","tool_call":{"shellToolCall":{"result":{"exit_code":0,"output":"ok"}}}}`
	return nil
}
```
