## Preconditions
- These tests exercise `crush.UnwrapEvent` (or its exported equivalent) directly.
- No server or external dependency needed.
- Each leaf provides a 3-level SSE JSON string as `req.SSEInput`.

## Steps
1. Set `req.Mode = "convert"`.
2. Set `req.SSEInput` with the specific 3-level SSE JSON.
3. Root `Run` calls `crush.UnwrapEvent` on `SSEInput`, marshals the result to JSON, and sets `resp.Output`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "convert"
	return nil
}
```
