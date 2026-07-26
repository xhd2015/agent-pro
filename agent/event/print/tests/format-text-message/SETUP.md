## Preconditions
- A text message event JSON line is supplied.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Line = `{"type":"text","part":{"id":"2","type":"text","text":"Hello world"}}`
	return nil
}
```
