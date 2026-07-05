# Scenario

**Feature**: tail emits matching tool_call_id on tool events

```
tool_call + tool_call_update -> TailUpdatesFromOffset -> both tool events share tool_call_id
```

## Preconditions

- One `tool_call` pending and one `tool_call_update` completed for `call_read_1`.

## Steps

1. Provide user chunk, tool pair, and `turn_completed` wire lines.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("read file"),
		acpToolCall("call_read_1", "read", "README.md"),
		acpToolCallUpdate("call_read_1", "completed", "file contents"),
		acpTurnCompleted(),
	}
	return nil
}
```