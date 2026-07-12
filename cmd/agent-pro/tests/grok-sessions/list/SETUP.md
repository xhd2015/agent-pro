# Scenario

**Feature**: list Grok sessions from synthetic GROK_HOME (with optional grep)

```
# walk GROK_HOME/sessions/*/uuid/summary.json and parse metadata
sessions.List | ListWithGrep(grokHome, limit, pattern) -> sessions / matches

# format as table; with grep, append indented hit lines (optional color)
[]Session -> FormatListTable | FormatListTableWithHits(now, color) -> terminal text
```

## Preconditions

- This branch tests the `list` operation only (including `list/grep/*`).
- Session fixtures are created by descendant leaf Setup functions.
- Grep leaves set `req.Grep` and optionally `req.Color`; classic list leaves leave them empty.

## Steps

1. Set `req.Operation = "list"`.
2. Leaf Setup writes `summary.json` / `chat_history.jsonl` as needed.
3. `Run` calls `List`+`FormatListTable` when Grep is empty, else
   `ListWithGrep`+`FormatListTableWithHits`.

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