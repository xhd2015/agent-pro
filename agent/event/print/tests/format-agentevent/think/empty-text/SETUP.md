## Preconditions
- A think event with empty text.

## Steps
1. Set `req.Text = ""` (explicit empty).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Text = ""
	return nil
}
```
