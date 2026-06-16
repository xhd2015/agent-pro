## Preconditions
- A think event with reasoning text.

## Steps
1. Set `req.Text = "reasoning about the problem..."`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Text = "reasoning about the problem..."
	return nil
}
```
