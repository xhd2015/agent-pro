# Scenario

**Bug**: follow-up runner prompt must include prior user content (e.g. "hi")

```
events [user:hi, assistant:hello] + "what did I ask?" -> built prompt contains "hi"
```

## Preconditions

- Prior turn completed with user `hi` and assistant `hello`.

## Steps

1. Set `PriorEvents` to one user and one assistant message.
2. Set follow-up `NewPrompt` to `what did I ask?`.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.PriorEvents = []types.AgentEvent{
		msgEvent("user", "hi"),
		msgEvent("assistant", "hello"),
	}
	req.NewPrompt = "what did I ask?"
	return nil
}
```