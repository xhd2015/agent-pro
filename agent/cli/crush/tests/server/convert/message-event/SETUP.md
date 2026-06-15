## Preconditions
- Input is a valid 3-level SSE line with outer type `"message"`.
- Inner payload contains a complete `MessagePayload` with `id`, `role`, `session_id`, `parts`.

## Steps
1. Set `SSEInput` to a 3-level message event with text and reasoning parts.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SSEInput = `{"type":"message","payload":{"type":"updated","payload":{"id":"msg_abc123","role":"assistant","session_id":"sess_xyz789","parts":[{"type":"text","data":{"text":"Hello, world!"}},{"type":"reasoning","data":{"thinking":"Let me think..."}}],"model":"deepseek-v4-flash","provider":"deepseek"}}}`
	return nil
}
```
