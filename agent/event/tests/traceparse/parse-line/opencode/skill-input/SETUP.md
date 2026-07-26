# Scenario

**Feature**: traceparse leaf `parse-line/opencode/skill-input`

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
	req.RawLine = `{"type":"tool_use","part":{"type":"tool","tool":"Skill","state":{"status":"completed","input":{"name":"git-fetch","arguments":"skill install --general-agents"}}}}`
	return nil
}
```
