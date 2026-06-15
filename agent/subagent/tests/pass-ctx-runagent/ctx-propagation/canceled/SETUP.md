## Preconditions
- `req.CancelCtx` is set to `true` so that the context passed to `runAgent` is already canceled.
- The agent runner (`fake-codex`) must be findable via `AGENT_RUNNER_FAKE_CODEX_PATH`.

## Steps
1. Set `req.CancelCtx = true`.
2. `Run` calls `runAgent` with the canceled context.
3. Verify the returned error is non-nil.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.CancelCtx = true
    req.Prompt = "test"
    return nil
}
```
