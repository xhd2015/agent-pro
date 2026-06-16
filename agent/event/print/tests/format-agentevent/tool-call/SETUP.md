## Preconditions
- The event Type is `tool_call`. The tool name is "bash" (shared default).

## Steps
1. Set `req.Type = types.ActionToolCall`.
2. Set `req.Tool = "bash"`.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Type = types.ActionToolCall
	req.Tool = "bash"
	return nil
}
```
