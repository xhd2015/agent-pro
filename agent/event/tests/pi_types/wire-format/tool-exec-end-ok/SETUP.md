## Preconditions
- Successful tool_execution_end has `result` and `isError: false`.

## Steps
1. Parse a tool_execution_end event with a successful result.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"tool_execution_end","toolCallId":"call_1","toolName":"bash","result":"file.txt\nfile2.txt","isError":false}`
	return nil
}
```
