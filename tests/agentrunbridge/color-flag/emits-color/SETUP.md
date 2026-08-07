# Scenario

**Feature**: BuildArgs emits `--color` when RunOpts.Color is true

```
RunOpts{Open, Color:true, Prompt}
  -> … --open --color -- <prompt>
```

## Steps

1. Set session, open profile flags, Color true.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-color-1"
	req.Prompt = "with color"
	req.AgentRunner = "grok-tty"
	req.AutoSendOrResume = true
	req.NewTerminal = true
	req.Open = true
	req.Color = true
	return nil
}
```
