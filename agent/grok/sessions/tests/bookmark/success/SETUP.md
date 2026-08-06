# Scenario

**Feature**: successful bookmark pin / list / show / remove / format

```
# live session on disk under injectable grokHome
writeBookmarkSession -> BookmarkGrok / List / Get / Remove / Format*
  -> store written or views returned without error
```

## Preconditions

- Grok session fixture exists unless a leaf only formats pure in-memory views.
- Agent-pro home is empty of store unless leaf preseeds catalog JSON.
- Leaves under this branch expect `resp.Err == nil` for the primary op.

## Steps

1. Default Op remains for leaves to override.
2. Leaf seeds session and/or store, sets pin opts or filters.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Success leaves configure fixtures; ensure homes from root remain.
	if req.AgentProHome == "" || req.GrokHome == "" {
		t.Fatal("expected AgentProHome and GrokHome from root Setup")
	}
	return nil
}
```
