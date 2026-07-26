# Scenario

**Feature**: tool_call wire emits ActionToolCall with pending status

```
tool_call{toolCallId, kind, title} -> ActionToolCall tool_call_id status=pending
```

## Preconditions
- One `tool_call` line with id `call_read_1`.

## Steps
1. Provide `acpToolCall("call_read_1", "read", "README.md")`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{acpToolCall("call_read_1", "read", "README.md")}
	return nil
}
```
