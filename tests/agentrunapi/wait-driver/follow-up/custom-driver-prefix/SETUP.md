# Scenario

**Feature**: custom DriverBinary + DriverArgsPrefix (spl helper shape)

```
BuildFollowUpCommand(
  DriverBinary=/usr/local/bin/spl-helper,
  DriverArgsPrefix=[local-bot, agent-exec],
  Open+session+prompt)
  -> line contains binary + prefix tokens + run
  -> no --new-terminal
```

## Steps

1. Set custom driver path and prefix tokens.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DriverBinary = "/usr/local/bin/spl-helper"
	req.DriverArgsPrefix = []string{"local-bot", "agent-exec"}
	req.SessionID = "sess-fu-custom"
	req.Prompt = "custom driver"
	req.AgentRunner = "grok-tty"
	req.Open = true
	return nil
}
```
