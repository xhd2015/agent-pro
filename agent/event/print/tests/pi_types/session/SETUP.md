## Preconditions
- A `session` event is supplied (should be skipped/hidden).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Line = `{"type":"session","id":"sess_1"}`
	return nil
}
```
