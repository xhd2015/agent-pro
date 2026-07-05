# Scenario

**Feature**: completed tool pair roundtrips

```
tool_call + tool_call_update -> tool_call_id + status + Output preserved
```

## Preconditions
- Tool call and completed update in one turn.

## Steps
1. Seed full tool pair wire lines.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("read file"),
		acpToolCall("call_read_1", "read", "README.md"),
		acpToolCallUpdate("call_read_1", "completed", "package main"),
		acpTurnCompleted(),
	}
	return nil
}
```
