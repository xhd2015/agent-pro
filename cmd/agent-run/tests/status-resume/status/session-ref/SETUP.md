# Scenario

**Feature**: status accepts `runner/session_id` unambiguous refs

```
seed meta under grok-tty/<id>
  -> agent-run status grok-tty/<id> -> same multi-layer view as bare id
```

## Steps

1. Leaf seeds unique meta and runs with `runner/session` form.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Grouping for runner/session_id ref form leaves.
	req.Runner = defaultRunner
	return nil
}
```
