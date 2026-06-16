## Preconditions
- The converter router receives crush-native JSON events.

## Steps
1. Set `req.AgentRunner = "crush"`.
2. Feed crush events: message with text part, message with tool_call part.
3. Call `convert.ConvertRawLine`.
4. Verify AgentEvent output.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.AgentRunner = "crush"
    req.RawJSON = `[{"type":"message","payload":"{\"id\":\"m1\",\"role\":\"assistant\",\"session_id\":\"s1\",\"parts\":[{\"type\":\"text\",\"data\":\"{\\\"text\\\":\\\"Crush response\\\"}\"}]}"},{"type":"message","payload":"{\"id\":\"m2\",\"role\":\"assistant\",\"session_id\":\"s1\",\"parts\":[{\"type\":\"tool_call\",\"data\":\"{\\\"id\\\":\\\"t1\\\",\\\"name\\\":\\\"read\\\",\\\"input\\\":{\\\"path\\\":\\\"/tmp/test\\\"}}\"}]}"}]`
    return nil
}
```
