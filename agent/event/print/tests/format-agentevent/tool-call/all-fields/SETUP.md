## Preconditions
- A tool_call event with all optional fields populated.

## Steps
1. Set `req.Text = "Searching..."`.
2. Set `req.Output = "found 3 matches"`.
3. Set `req.Changes = []types.FileChange{{Path: "a.txt", Kind: "modify"}}`.
4. Set `req.ExitCode` to a pointer to 2 (error, to also show FAILED).

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	exitErr := 2
	req.Text = "Searching..."
	req.Output = "found 3 matches"
	req.Changes = []types.FileChange{{Path: "a.txt", Kind: "modify"}}
	req.ExitCode = &exitErr
	return nil
}
```
