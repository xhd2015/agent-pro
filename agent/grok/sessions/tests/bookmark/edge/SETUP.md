# Scenario

**Feature**: hybrid orphan list, empty store, multi-runner catalog edges

```
# orphan: pin then remove session dir -> list Orphaned + warning
# missing store: ListBookmarks -> empty
# preseed codex+grok: list shows both; ambiguous show needs runner
```

## Preconditions

- Edges exercise hybrid enrich, empty catalog, and multi-runner identity rules.
- Still L2 injectable homes only.

## Steps

1. Leaf seeds the edge fixture.
2. `Run` lists/shows as configured.
3. Assert Orphaned, warnings, multi-runner presence, or empty table.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.GrokHome == "" || req.AgentProHome == "" {
		t.Fatal("expected homes from root Setup")
	}
	return nil
}
```
