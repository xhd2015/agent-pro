# Scenario

**Feature**: Model set, Effort empty → only `--model`

```
BuildFollowUpCommand(Model=o3, Effort="", Open, …)
  -> --model o3
  -> no --model-reasoning-effort
```

## Steps

1. Set Model=`o3`, Effort empty.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-fu-model-only"
	req.Prompt = "model only"
	req.Model = fixtureModel // "o3"
	req.ModelReasoningEffort = ""
	req.Open = true
	return nil
}
```
