## Preconditions
- `req.CancelCtx` is `false` (default), so the context is a normal `context.Background()`.
- The agent runner (`fake-codex`) responds with output.

## Steps
1. Set `req.Prompt` to a simple prompt.
2. `Run` calls `runAgent` with the normal context.
3. Verify no error and output is non-empty.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Prompt = "write hello world"
    return nil
}
```
