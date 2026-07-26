# Scenario

**Feature**: failed tool pair roundtrips

```
tool_call + failed update -> status=failed preserved
```

## Preconditions
- Failed tool_call_update.

## Steps
1. Seed tool pair with failed status.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("run false"),
		acpToolCall("call_exec_1", "execute", "false"),
		acpToolCallUpdate("call_exec_1", "failed", "exit code 1"),
		acpTurnCompleted(),
	}
	return nil
}
```
