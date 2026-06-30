# Scenario

**Feature**: `sessions` list mode (no session positional)

```
agent-run sessions [--json] -> scan AGENT_RUN_HOME/sessions -> stdout list or JSON
```

## Preconditions

- No `<runner>/<session_id>` positional argument.

## Steps

1. Prefix `req.Args` with `sessions`.
2. Leaf adds flags such as `--json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"sessions"}
	return nil
}
```