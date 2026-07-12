# Scenario

**Feature**: compound `runner/session_id` ref is rejected (Q5)

```
agent-run sessions fake-codex/web_test123 --print -> exit 1; invalid session reference
```

## Preconditions

- Bare session id only; any ref containing `/` is invalid.
- Session may or may not exist; ref validation fails first.

## Steps

1. Optionally seed a flat session so failure is purely ref format.
2. Run print with compound positional `fake-codex/web_test123`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	store := openAgentStore(t, req)
	seedSessionMeta(t, store, printSessionID, "finished")
	req.Args = []string{"sessions", printRunner + "/" + printSessionID, "--print"}
	return nil
}
```
