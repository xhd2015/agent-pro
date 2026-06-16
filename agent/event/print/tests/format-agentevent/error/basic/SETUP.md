## Preconditions
- An error event with text.

## Steps
1. Set `req.Text = "boom"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Text = "boom"
	return nil
}
```
