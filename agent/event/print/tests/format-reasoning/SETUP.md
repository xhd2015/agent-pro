## Preconditions
- A reasoning event JSON line is supplied.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Line = `{"type":"reasoning","part":{"id":"1","type":"reasoning","text":"Let me think about this carefully"}}`
	return nil
}
```
