# Scenario

**Feature**: tool_call_update with completed status sets Output

```
tool_call + tool_call_update{status:completed} -> ActionToolCall with Output status=completed
```

## Preconditions
- `tool_call` followed by `tool_call_update` with same id.

## Steps
1. Provide tool_call then completed tool_call_update with output text.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{
		acpToolCall("call_read_1", "read", "README.md"),
		acpToolCallUpdate("call_read_1", "completed", "package main"),
	}
	return nil
}
```
