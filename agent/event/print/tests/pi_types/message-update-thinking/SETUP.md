## Preconditions
- A `message_update` event with thinking delta and assistant role is supplied.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Line = `{"type":"message_update","message":{"role":"assistant","content":[{"type":"thinking","thinking":"hmm"}]},"assistantMessageEvent":{"type":"thinking_delta","delta":"hmm"}}`
	return nil
}
```
