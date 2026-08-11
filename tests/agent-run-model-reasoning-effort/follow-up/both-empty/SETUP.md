# Scenario

**Feature**: both empty → neither flag; no agent-pro defaults

```
BuildFollowUpCommand(Model="", Effort="", Open, …)
  -> no --model
  -> no --model-reasoning-effort
  -> does not invent gpt-5.6-luna or max
```

## Steps

1. Leave Model and Effort empty.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-fu-both-empty"
	req.Prompt = "both empty"
	req.Model = ""
	req.ModelReasoningEffort = ""
	req.Open = true
	return nil
}
```
