## Preconditions
- A tool_call event with both text and output.

## Steps
1. Set `req.Text = "Running command..."`.
2. Set `req.Output = "command result"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Text = "Running command..."
	req.Output = "command result"
	return nil
}
```
