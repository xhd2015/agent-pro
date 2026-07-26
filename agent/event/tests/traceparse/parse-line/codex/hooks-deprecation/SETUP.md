# Scenario

**Feature**: traceparse leaf `parse-line/codex/hooks-deprecation`

```
trace JSONL -> adapter registry -> parsed event JSON
```

## Preconditions
- Mode and inputs are set for this leaf.

## Steps
1. Load warning JSONL fixture into `req.RawLine`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Fixture inlined: warning-line.jsonl is gitignored (*.jsonl).
	req.RawLine = `{"type":"item.completed","item":{"id":"item_0","type":"error","message":"` + "`[features].codex_hooks` is deprecated. Use `[features].hooks` instead." + `"}}`
	return nil
}
```