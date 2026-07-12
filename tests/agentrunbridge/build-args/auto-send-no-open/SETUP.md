# Scenario

**Feature**: AutoSendOrResume without Open

```
BuildArgs(AutoSendOrResume=true, Open=false)
  -> includes --auto-send-or-resume; omits --open
```

## Preconditions

- Session present; auto-send true; open false; new-terminal optional false.

## Steps

1. Set session, auto-send, prompt; Open remains false.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "sess-autosend"
	req.Prompt = "auto only"
	req.AgentRunner = "grok-tty"
	req.AutoSendOrResume = true
	req.Open = false
	req.NewTerminal = false
	return nil
}
```
