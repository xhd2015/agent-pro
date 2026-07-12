# Scenario

**Feature**: empty resume id keeps continuation history wrap for unbound multi-turn

```
resumeID="" + prior[user:hi, assistant:hello] + "follow-up please"
  -> ResolveRunnerPrompt
  -> BuildContinuationPrompt(prior, "follow-up please")
  -> contains "Previous conversation", "hi", and follow-up
```

## Preconditions

- `ResumeID` is empty (no provider-native resume; e.g. web multi-turn with fake-codex).
- Same prior transcript as the skip-history sibling: user `hi`, assistant `hello`.
- Same new prompt `follow-up please` so the only variable is resume id.

## Steps

1. Set empty `ResumeID`, prior history, and follow-up prompt.
2. Expect inject text to match unbound continuation behavior.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.ResumeID = ""
	req.NewPrompt = "follow-up please"
	req.PriorEvents = []types.AgentEvent{
		msgEvent("user", "hi"),
		msgEvent("assistant", "hello"),
	}
	return nil
}
```
