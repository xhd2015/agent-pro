## Preconditions
- A `message_start` event with assistant role and text content is supplied.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Line = `{"type":"message_start","message":{"role":"assistant","content":[{"type":"text","text":"Hello"}]}}`
	return nil
}
```
