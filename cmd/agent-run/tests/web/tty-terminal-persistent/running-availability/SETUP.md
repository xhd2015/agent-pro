# Scenario

**Bug**: running web-created tty sessions report terminal unavailable until the turn finishes

```
web-created codex-tty running session + live tty registry entry
  -> /api/agent-run/sessions/codex-tty/web_*/terminal
  -> available:true before assistant response finishes
```

## Preconditions

- Descendant leaves create real web `codex-tty` sessions through the HTTP API.
- The fake `codex` binary keeps the turn running long enough to probe terminal
  availability before the response finishes.

## Steps

1. Create a running `codex-tty` session through the web API.
2. Wait for a live tty registry entry.
3. Probe terminal state through the web chat session id while the turn is still running.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Status = "running"
	return nil
}
```
