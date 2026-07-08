# Scenario

**Bug**: P1 — delayed tool completion and assistant must reach `events.jsonl` under keep-tty

```
pre-seed user + think + pending tool_call
  -> 900ms later: tool_call_update(completed) + agent_message + turn_completed
  -> events.jsonl has completed tool, assistant marker, done after assistant
```

## Steps

1. Fixed session id `chat_tail_p1`.
2. Schedule completion append 900ms after grok-tty session id appears.
3. Poll `events.jsonl` until assistant marker + ordering satisfied.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "delayed-tool-completion-in-events"
	req.SessionID = "chat_tail_p1"
	req.CompletionDelay = 3000 * time.Millisecond
	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{{
		Delay: req.CompletionDelay,
		Lines: completionAppendLines(),
	}}
	return nil
}
```