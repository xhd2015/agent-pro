# Scenario

**Feature**: multi-turn transcripts preserve chronological user lines in the prefix

```
two complete turns + third user prompt -> first topic appears before second topic
```

## Preconditions

- At least two prior user messages with distinct text.

## Steps

1. Seed four message events (user/assistant × 2).
2. Third user prompt is a new follow-up.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.PriorEvents = []types.AgentEvent{
		msgEvent("user", "first-topic"),
		msgEvent("assistant", "reply-one"),
		msgEvent("user", "second-topic"),
		msgEvent("assistant", "reply-two"),
	}
	req.NewPrompt = "summarize both topics"
	return nil
}
```