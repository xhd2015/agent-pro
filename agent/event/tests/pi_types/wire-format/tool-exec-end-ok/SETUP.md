# Scenario

**Feature**: Successful tool_execution_end has `result` and `isError: false`

## Preconditions
- Successful tool_execution_end has `result` and `isError: false`.

## Steps
1. Parse a tool_execution_end event with a successful result.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"tool_execution_end","toolCallId":"call_1","toolName":"bash","result":"file.txt\nfile2.txt","isError":false}`
	return nil
}
```
