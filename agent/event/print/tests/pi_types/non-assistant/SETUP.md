## Preconditions
- A `message_update` event with non-assistant role (user) is supplied.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Line = `{"type":"message_update","message":{"role":"user","content":[{"type":"text","text":"hello"}]}}`
	return nil
}
```
