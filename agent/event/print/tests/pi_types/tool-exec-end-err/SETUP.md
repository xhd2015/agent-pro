## Preconditions
- A `tool_execution_end` event with failed bash result is supplied.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Line = `{"type":"tool_execution_end","toolCallId":"c1","toolName":"bash","result":"not found","isError":true}`
	return nil
}
```
