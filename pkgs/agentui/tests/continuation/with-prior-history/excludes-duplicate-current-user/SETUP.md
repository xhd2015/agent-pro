# Scenario

**Feature**: current user turn appended to events must not appear twice in the prefix

```
[user:hi, assistant:hello, user:what did I ask?] + same new prompt -> one "what did I ask?" in output
```

## Preconditions

- Store may already have appended the follow-up user line before `Run` reads history.

## Steps

1. Prior events end with the same text as `NewPrompt`.
2. `BuildContinuationPrompt` excludes that trailing user line from the history block.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.NewPrompt = "what did I ask?"
	req.PriorEvents = []types.AgentEvent{
		msgEvent("user", "hi"),
		msgEvent("assistant", "hello"),
		msgEvent("user", req.NewPrompt),
	}
	return nil
}
```