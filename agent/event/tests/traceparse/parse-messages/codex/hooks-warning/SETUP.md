# Scenario

**Feature**: traceparse leaf `parse-messages/codex/hooks-warning`

```
trace lines[] + created_at -> message aggregator -> Config Warning tool call
```

## Preconditions
- Mode and inputs are set for this leaf.

## Steps
1. Load warning JSONL fixture into `req.RawLines`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Fixture inlined: warning-line.jsonl is gitignored (*.jsonl).
	req.RawLines = []string{
		`{"type":"item.completed","item":{"id":"item_0","type":"error","message":"` + "`[features].codex_hooks` is deprecated. Use `[features].hooks` instead." + `"}}`,
	}
	req.CreatedAt = "2026-05-25T18:26:22.524536+08:00"
	return nil
}
```