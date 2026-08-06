# Scenario

**Feature**: missing catalog file lists as empty (no error)

```
# no session_bookmarks.json
ListBookmarks / FormatBookmarksTable
  -> empty views; table empty or "No bookmarks" style
```

## Preconditions

- Store file absent.
- Op=format-list (also exercises list empty path).

## Steps

1. Do not write store.
2. Op=format-list.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Ensure store absent.
	assertPathMissing(t, storePath(req.AgentProHome))
	req.Op = "format-list"
	return nil
}
```
