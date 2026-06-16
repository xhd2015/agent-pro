## Preconditions
- The converter router receives opencode-native JSON with multiple event types.

## Steps
1. Set `req.AgentRunner = "opencode"`.
2. Feed two opencode events: reasoning and tool_use.
3. Call `convert.ConvertRawLine`.
4. Verify the resulting AgentEvent JSON contains correct `type`, `text`, `tool`, `tool_input` fields.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.AgentRunner = "opencode"
    req.RawJSON = `[{"type":"reasoning","timestamp":1718444400000,"part":{"id":"r1","type":"reasoning","text":"analyzing the request"}},{"type":"tool_use","timestamp":1718444401000,"part":{"id":"t1","type":"tool","callID":"t1","tool":"bash","state":{"input":{"command":"ls"},"output":"file1.txt\nfile2.txt","exit_code":0,"status":"completed"}}}]`
    return nil
}
```
