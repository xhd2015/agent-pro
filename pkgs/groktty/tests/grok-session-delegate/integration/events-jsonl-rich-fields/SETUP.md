# Scenario

**Feature**: tail events match grok_session.FromUpdatesJSONL semantics

```
full-turn wire fixture -> tail events ≡ grok_session.FromUpdatesJSONL (semantic)
```

## Preconditions

- Full turn with user, think, tool, assistant, and `turn_completed`.
- Parity requires groktty tail to delegate to grok_session converter.

## Steps

1. Provide the same full-turn wire lines as `tail/emits-turn-index-and-done`.

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