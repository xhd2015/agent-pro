# Scenario

**Feature**: empty DriverBinary defaults to agent-run

```
BuildFollowUpCommand(DriverBinary="", Open, SessionID, Prompt)
  -> line contains "agent-run"
```

## Steps

1. Leave DriverBinary empty; open profile fields set.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DriverBinary = ""
	req.SessionID = "sess-fu-default"
	req.Prompt = "open me"
	req.AgentRunner = "grok-tty"
	req.Open = true
	req.Detach = false
	return nil
}
```
