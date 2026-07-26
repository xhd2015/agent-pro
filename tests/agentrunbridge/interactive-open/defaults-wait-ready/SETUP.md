# Scenario

**Feature**: minimal InteractiveOpen opts launch open profile and wait ready

```
RunInteractiveOpen(SessionID, Prompt) -> launch open argv -> poll ready -> ok
```

## Preconditions

- Only SessionID + Prompt required.
- Status poll returns ready on first call (grouping default).

## Steps

1. Set SessionID and Prompt; leave AgentRunner empty (default grok-tty).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "interactive_open"
	req.SessionID = "sess-io-default"
	req.Prompt = "interactive hello"
	req.AgentRunner = ""
	return nil
}
```
