# Scenario

**Feature**: Error tool_execution_end has `isError: true` and may contain error details in result

## Preconditions
- Error tool_execution_end has `isError: true` and may contain error details in result.

## Steps
1. Parse a tool_execution_end event with isError:true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"tool_execution_end","toolCallId":"call_2","toolName":"read","result":"file not found","isError":true}`
	return nil
}
```
