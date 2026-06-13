## Preconditions
- An error event JSON line is supplied.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Line = `{"type":"error","error":{"name":"Error","data":{"message":"something broke"}}}`
	return nil
}
```
