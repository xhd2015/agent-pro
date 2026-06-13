## Preconditions
- This branch tests the `grep` tool execution.
- The mock config must contain a tool_use event with `"tool":"grep"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_TOOL=grep")
    return nil
}
```
