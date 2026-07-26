# Scenario

**Feature**: `ListSessions` returns all sessions regardless of runner field

```
CreateSession(s1, runner=fake-opencode) + CreateSession(s2, runner=fake-codex)
-> ListSessions() -> both sessions
```

## Preconditions

- Two sessions exist with different `meta.runner` values and distinct bare ids.
- List has no runner filter argument.

## Steps

1. Set `req.Action = "list_all"`.
2. Create one session for `fake-opencode` and one for `fake-codex`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "list_all"
	req.Runner = "fake-opencode"
	req.SessionID = "sess_opencode"
	req.OtherRunner = "fake-codex"
	req.OtherSessID = "sess_codex"
	return nil
}
```
