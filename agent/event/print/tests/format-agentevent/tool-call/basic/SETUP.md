## Preconditions
- A basic tool_call event with output.

## Steps
1. Set `req.Output = "file1.txt"`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Output = "file1.txt"
	return nil
}
```
