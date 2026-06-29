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
	"os"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	b, err := os.ReadFile("warning-line.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	req.RawLine = strings.TrimSpace(string(b))
	return nil
}
```