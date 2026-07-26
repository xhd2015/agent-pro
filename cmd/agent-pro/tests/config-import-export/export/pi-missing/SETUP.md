## Preconditions
- No pi directory exists on disk.
- The operation has been set to "export".

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "export"
	req.Agent = "pi"
	return nil
}
```
