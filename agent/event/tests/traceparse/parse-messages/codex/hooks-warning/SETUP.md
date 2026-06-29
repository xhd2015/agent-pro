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
	"os"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	b, err := os.ReadFile("warning-line.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	req.RawLines = []string{strings.TrimSpace(string(b))}
	req.CreatedAt = "2026-05-25T18:26:22.524536+08:00"
	return nil
}
```