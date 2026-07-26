# Scenario

**Feature**: traceparse leaf `print-integration/format-trace-line-codex`

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
	req.RawLine = `{"type":"item.completed","item":{"type":"agent_message","text":"Here is the answer."}}`
	return nil
}
```
