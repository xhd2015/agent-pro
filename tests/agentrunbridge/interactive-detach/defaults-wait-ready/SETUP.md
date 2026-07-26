# Scenario

**Feature**: minimal InteractiveDetach opts launch detach profile and wait ready

```
RunInteractiveDetach(SessionID, Prompt) -> launch detach argv -> poll ready -> ok
```

## Preconditions

- Only SessionID + Prompt required.
- Status poll returns ready on first call (grouping default).

## Steps

1. Set SessionID and Prompt; leave AgentRunner empty (default grok-tty).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "interactive_detach"
	req.SessionID = "sess-id-default"
	req.Prompt = "interactive detach hello"
	req.AgentRunner = ""
	return nil
}
```
