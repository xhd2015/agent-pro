## Preconditions
- This branch tests the `write` tool execution.
- The mock config must contain a tool_use event with `"tool":"write"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_TOOL=write")
    return nil
}
```
