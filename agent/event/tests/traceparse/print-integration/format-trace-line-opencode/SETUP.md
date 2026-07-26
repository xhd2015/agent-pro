# Scenario

**Feature**: traceparse leaf `print-integration/format-trace-line-opencode`

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
	req.RawLine = `{"type":"text","sessionID":"sess-123","timestamp":1700000000000,"part":{"id":"part-1","type":"text","text":"I have completed the task."}}`
	return nil
}
```
