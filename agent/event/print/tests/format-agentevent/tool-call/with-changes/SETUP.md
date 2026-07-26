## Preconditions
- A tool_call event with file changes.

## Steps
1. Override `req.Tool = "write"` (to test EDIT icon).
2. Set `req.Changes = []types.FileChange{{Path: "src/main.go", Kind: "create"}}`.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Tool = "write"
	req.Changes = []types.FileChange{{Path: "src/main.go", Kind: "create"}}
	return nil
}
```
