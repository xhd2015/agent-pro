# Scenario

**Feature**: `ListSessions` returns only sessions for the requested runner

```
CreateSession(fake-opencode/s1) + CreateSession(fake-codex/s2) -> ListSessions(fake-opencode) -> [s1]
```

## Preconditions

- Two sessions exist under different runners.
- List query targets a single runner.

## Steps

1. Set `req.Action = "list_by_runner"`.
2. Create one session for `fake-opencode` and one for `fake-codex`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "list_by_runner"
	req.Runner = "fake-opencode"
	req.SessionID = "sess_opencode"
	req.OtherRunner = "fake-codex"
	req.OtherSessID = "sess_codex"
	return nil
}
```