# Scenario

**Feature**: both Model and Effort set → both flags on follow-up line

```
BuildFollowUpCommand(Model=o3, Effort=high, Open, …)
  -> --model o3 (or --model=o3)
  -> --model-reasoning-effort high (or equals form)
```

## Steps

1. Set Model=`o3`, ModelReasoningEffort=`high`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-fu-both"
	req.Prompt = "both set"
	req.Model = fixtureModel       // "o3"
	req.ModelReasoningEffort = fixtureEffort // "high"
	req.Open = true
	return nil
}
```
