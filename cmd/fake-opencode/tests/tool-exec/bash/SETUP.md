## Preconditions
- This branch tests the `bash` tool execution.
- The mock config must contain a tool_use event with `"tool":"bash"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_TOOL=bash")
    return nil
}
```
