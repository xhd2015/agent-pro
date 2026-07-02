# Scenario

**Feature**: stale tty registry is not advertised as attachable

```
codex-tty session + registry listen_addr on closed port -> GET /terminal -> available false
```

## Preconditions

- Resolver checks registry reachability before reporting attach availability.

## Steps

1. Write tty session metadata.
2. Write registry entry pointing at an unused localhost address.
3. Fetch terminal status.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "codex-tty"
	req.SessionID = "stale-terminal-session"
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running")
	writeTTYRegistryFixture(t, req, req.Runner, req.SessionID, unusedLocalAddr(t))
	req.HTTPPath = terminalStatusPath(req.Runner, req.SessionID)
	return nil
}
```
