# Scenario

**Feature**: empty sessions tree yields empty list and friendly table phrase

```
empty GROK_HOME/sessions
  -> ListWithOptions -> []
  -> FormatListTable -> "No sessions found"
```

## Preconditions

- No summary.json under sessions.
- WantFormat true.

## Steps

1. Leave sessions tree empty (root only created sessions dir).
2. Set Limit=20, WantFormat=true.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20
	req.WantFormat = true
	return nil
}
```
