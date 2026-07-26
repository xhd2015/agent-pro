# Scenario

**Feature**: list OpenCode sessions from storage/session JSON files

```
# walk storage/session/{projectID}/ses_*.json
sessions.List(dataDir, limit) -> []Session sorted by time.updated desc

# default table formatter shows unified grok-shaped columns
FormatListTable(sessions, dataDir, now) -> terminal table
```

## Preconditions

- Default operation is `list` when `req.Operation` is empty.
- `req.Now` is fixed by root Setup for relative LAST ACTIVE times.

## Steps

1. Leaf Setup writes one or more session JSON fixtures.
2. Run calls `sessions.List` and `FormatListTable`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "list"
	return nil
}
```