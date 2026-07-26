## Preconditions
- A whitespace-only line is supplied.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Line = `   `
	return nil
}
```
