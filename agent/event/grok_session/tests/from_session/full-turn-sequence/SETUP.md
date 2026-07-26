# Scenario

**Feature**: complete single turn with all update kinds

```
user + think + tool_call + tool_call_update + assistant + turn_completed
```

## Preconditions
- Representative full-turn wire sequence.

## Steps
1. Provide full-turn wire lines ending in turn_completed.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("persist events"),
		acpThoughtChunk("planning ls output"),
		acpToolCall("call_persist", "execute", "ls"),
		acpToolCallUpdate("call_persist", "completed", "agent\nagents"),
		acpAssistantChunk("PERSIST_ASSISTANT_MARKER"),
		acpTurnCompleted(),
	}
	return nil
}
```
