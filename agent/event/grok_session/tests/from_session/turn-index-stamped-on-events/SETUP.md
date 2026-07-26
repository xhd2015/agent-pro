# Scenario

**Feature**: all events in one turn share the same turn_index

```
user + think + tool + assistant + turn_completed -> all events turn_index=0
```

## Preconditions
- Full single-turn wire sequence before turn_completed.

## Steps
1. Provide user, thought, tool_call, tool_call_update, assistant, turn_completed.

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
