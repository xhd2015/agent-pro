## Preconditions
- A non-JSON plain text line is supplied.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Line = `this is just some text, not JSON`
	return nil
}
```
