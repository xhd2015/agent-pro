# Scenario

**Feature**: Stateless run — no session id flag

```
BuildArgs(Stateless=true, prompt)
  -> run [optional runner flags] <prompt>
  # no --session-id
```

## Preconditions

- `Stateless=true`; session id may be set but must not appear in argv.
- Empty AgentRunner (omit runner flag).

## Steps

1. Set Stateless, prompt; clear runner.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Stateless = true
	req.SessionID = "should-not-appear"
	req.Prompt = "stateless prompt"
	req.AgentRunner = ""
	return nil
}
```
