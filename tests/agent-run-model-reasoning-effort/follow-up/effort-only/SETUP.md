# Scenario

**Feature**: Effort set, Model empty → only `--model-reasoning-effort`

```
BuildFollowUpCommand(Model="", Effort=max, Open, …)
  -> --model-reasoning-effort max
  -> no --model
```

## Steps

1. Set Model empty, Effort=`max`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-fu-effort-only"
	req.Prompt = "effort only"
	req.Model = ""
	req.ModelReasoningEffort = fixtureEffortMax // "max"
	req.Open = true
	return nil
}
```
