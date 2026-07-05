# Scenario

**Feature**: full turn emits turn_index and ActionDone

```
user + think + tool + assistant + turn_completed -> turn_index=0 on all events; ActionDone present
```

## Preconditions

- Representative full-turn wire sequence ending in `turn_completed`.

## Steps

1. Provide full-turn wire lines (same shape as grok_session full-turn-sequence).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
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