# Scenario

**Feature**: `sessions` list mode (no session positional)

```
agent-run sessions [--json] [--limit N]
# sort: updated_at desc, then created_at, then session_id
# default limit 10; --limit 0 = all
# human columns: SESSION_ID RUNNER STATUS UPDATED
```

## Preconditions

- No `<session_id>` positional argument (list mode).
- Sessions live under flat `AGENT_RUN_HOME/sessions/<id>/`.

## Steps

1. Prefix `req.Args` with `sessions`.
2. Leaf adds flags such as `--json` / `--limit N` and seeds sessions as needed.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"sessions"}
	return nil
}
```
