## Preconditions
- This branch tests the `read` tool execution.
- The mock config must contain a tool_use event with `"tool":"read"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_TOOL=read")
    return nil
}
```
