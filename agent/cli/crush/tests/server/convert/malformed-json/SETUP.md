## Preconditions
- Input is not valid JSON at all (garbage string).
- The function must handle this gracefully.

## Steps
1. Set `SSEInput` to a non-JSON string.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SSEInput = `not valid json at all`
	return nil
}
```
