## Preconditions
- A `message_update` event with text delta and assistant role is supplied.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Line = `{"type":"message_update","message":{"role":"assistant","content":[{"type":"text","text":"world"}]},"assistantMessageEvent":{"type":"text_delta","delta":"world"}}`
	return nil
}
```
