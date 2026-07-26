# Scenario

**Feature**: read after byte offset skips already-consumed events

```
AppendEvent x3 -> ReadEvents(0) -> offset N -> ReadEvents(N) -> 0 new events
```

## Preconditions

- Three events are appended sequentially.
- Second read uses the offset returned by the first read.

## Steps

1. Set `req.Action = "read_after_offset"`.
2. Append three events; `Run` reads from 0 to capture offset, then reads again.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "read_after_offset"
	req.SessionID = "sess_offset"
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, Text: "one"},
		{Type: types.ActionMessage, Text: "two"},
		{Type: types.ActionMessage, Text: "three"},
	}
	return nil
}
```