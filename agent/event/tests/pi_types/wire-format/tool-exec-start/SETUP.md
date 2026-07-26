# Scenario

**Feature**: tool_execution_start has `toolCallId`, `toolName`, and `args` fields

## Preconditions
- tool_execution_start has `toolCallId`, `toolName`, and `args` fields.

## Steps
1. Parse a tool_execution_start event.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"tool_execution_start","toolCallId":"call_1","toolName":"bash","args":{"command":"ls -la"}}`
	return nil
}
```
