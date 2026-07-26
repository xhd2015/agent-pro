# Scenario

**Feature**: traceparse leaf `summary/title-from-identifier`

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
	req.SubMode = "title"
	req.Identifier = "shell_tool_call"
	return nil
}
```
