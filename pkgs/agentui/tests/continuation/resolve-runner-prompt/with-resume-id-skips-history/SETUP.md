# Scenario

**Feature**: bound provider resume id skips history wrap even when prior transcript exists

```
resumeID="grok-sess-abc" + prior[user:hi, assistant:hello] + "follow-up please"
  -> ResolveRunnerPrompt -> "follow-up please"
```

## Preconditions

- `ResumeID` is non-empty (provider will be invoked with native resume).
- Prior events include at least one complete user/assistant turn (`hi` / `hello`).
- New user turn is a distinct follow-up (`follow-up please`).

## Steps

1. Set non-empty `ResumeID`, prior history, and new prompt.
2. Expect inject text to be the trimmed new prompt only — no `Previous conversation` dump.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.ResumeID = "grok-sess-abc"
	req.NewPrompt = "follow-up please"
	req.PriorEvents = []types.AgentEvent{
		msgEvent("user", "hi"),
		msgEvent("assistant", "hello"),
	}
	return nil
}
```
