# Scenario

**Feature**: session ref without `--print` is rejected

```
seed session -> sessions fake-codex/web_test123 (no --print) -> exit 1
```

## Preconditions

- `--print` is required when a `<runner>/<session_id>` positional is given.

## Steps

1. Seed any session so list-vs-print dispatch is unambiguous.
2. Run `agent-run sessions fake-codex/web_test123` without `--print`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	store := openAgentStore(t, req)
	seedSessionMeta(t, store, printRunner, printSessionID, "finished")
	req.Args = []string{"sessions", printRunner + "/" + printSessionID}
	return nil
}
```