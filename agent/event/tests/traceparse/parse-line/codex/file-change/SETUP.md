# Scenario

**Feature**: traceparse leaf `parse-line/codex/file-change`

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
	req.RawLine = `{"type":"item.completed","item":{"id":"item_8","type":"file_change","changes":[{"path":"/tmp/code-commits.json","kind":"add"}],"status":"completed"}}`
	return nil
}
```
