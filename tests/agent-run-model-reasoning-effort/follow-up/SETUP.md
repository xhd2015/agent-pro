# Scenario

**Feature**: pure `BuildFollowUpCommand` forwards Model / ModelReasoningEffort when set

```
FollowUpOpts{Model, ModelReasoningEffort, Open, SessionID, …}
  -> BuildFollowUpCommand
  -> shell-quoted line with opt-in flags only
```

## Steps

1. Set mode `follow_up`.
2. Leaf sets Model / Effort fixtures.
3. Assert flag presence/absence (prefix-safe for `--model` vs `--model-reasoning-effort`).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "follow_up"
	return nil
}
```
