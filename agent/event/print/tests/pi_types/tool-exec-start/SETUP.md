## Preconditions
- A `tool_execution_start` event with bash tool and command args is supplied.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Line = `{"type":"tool_execution_start","toolCallId":"c1","toolName":"bash","args":{"command":"ls -la"}}`
	return nil
}
```
