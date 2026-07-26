# Scenario

**Feature**: tool_call_update with failed status

```
tool_call + tool_call_update{status:failed} -> ActionToolCall status=failed
```

## Preconditions
- Completed tool_call_update with `status=failed`.

## Steps
1. Provide tool_call then failed tool_call_update.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{
		acpToolCall("call_exec_1", "execute", "false"),
		acpToolCallUpdate("call_exec_1", "failed", "exit code 1"),
	}
	return nil
}
```
