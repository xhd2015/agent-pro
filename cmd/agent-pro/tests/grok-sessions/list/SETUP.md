# Scenario

**Feature**: list Grok sessions from synthetic GROK_HOME

```
# walk GROK_HOME/sessions/*/uuid/summary.json and parse metadata
sessions.List(grokHome, limit) -> []Session sorted newest first

# format as table with relative last-active times
[]Session -> FormatListTable(now) -> terminal text
```

## Preconditions

- This branch tests the `list` operation only.
- Session fixtures are created by descendant leaf Setup functions.

## Steps

1. Set `req.Operation = "list"`.
2. Leaf Setup writes `summary.json` files as needed for the scenario.
3. `Run` calls `sessions.List` then `sessions.FormatListTable` with `req.Now`.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Operation = "list"
	sessionsRoot := filepath.Join(req.GrokHome, "sessions")
	if err := os.MkdirAll(sessionsRoot, 0755); err != nil {
		t.Fatalf("mkdir sessions root: %v", err)
	}
	return nil
}
```