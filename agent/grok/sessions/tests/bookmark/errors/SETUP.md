# Scenario

**Feature**: bookmark APIs fail closed on bad input or corrupt store

```
# missing session / empty id / missing bookmark / corrupt JSON
BookmarkGrok | ListBookmarks | GetBookmark | RemoveBookmark
  -> non-nil error; corrupt store not silent-wiped
```

## Preconditions

- Leaves under this branch expect a non-nil `resp.Err`.
- Pin unknown must not create a catalog file.
- Corrupt store body must remain unchanged after failed ops.

## Steps

1. Leaf configures the failure condition.
2. `Run` dispatches the op under test.
3. Assert error class and store side effects.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Errors leaves set their own SessionID / store state.
	if req.AgentProHome == "" {
		t.Fatal("expected AgentProHome from root Setup")
	}
	return nil
}
```
