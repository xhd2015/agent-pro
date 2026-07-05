# Scenario

**Feature**: tail emits grok_session.status on tool events

```
tool_call pending + tool_call_update completed -> status pending then completed
```

## Preconditions

- Same tool pair as emits-tool-call-id.

## Steps

1. Provide tool_call pending and tool_call_update completed for `call_read_1`.

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