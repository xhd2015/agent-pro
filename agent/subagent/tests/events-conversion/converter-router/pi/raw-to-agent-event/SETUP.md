## Preconditions
- The converter router receives pi-native JSON events.

## Steps
1. Set `req.AgentRunner = "pi"`.
2. Feed pi events: message_start, tool_execution_start.
3. Call `convert.ConvertRawLine`.
4. Verify AgentEvent output.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.AgentRunner = "pi"
    req.RawJSON = `[{"type":"message_start","message":{"role":"assistant","content":[{"type":"text","text":"Hello from pi"}]}},{"type":"tool_execution_start","toolCallId":"t1","toolName":"bash","args":{"command":"echo hi"}}]`
    return nil
}
```
