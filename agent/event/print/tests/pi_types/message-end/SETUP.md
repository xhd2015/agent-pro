## Preconditions
- A `message_end` event with assistant role and text content is supplied.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Line = `{"type":"message_end","message":{"role":"assistant","content":[{"type":"text","text":"Bye"}]}}`
	return nil
}
```
