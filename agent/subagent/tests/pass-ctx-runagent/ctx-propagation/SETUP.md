## Preconditions
- Tests that the `ctx` parameter added to `runAgent` propagates correctly:
  - When a normal (non-canceled) context is passed, the agent runs successfully.
  - When a pre-canceled context is passed, `runAgent` returns an error.

## Steps
1. Set `req.AgentRunner` to `"fake-codex"`.
2. Each leaf sets `req.CancelCtx` to `true` or `false` and configures prompt/script.
3. `Run` calls `runAgent(ctx, ...)` and returns the result.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.AgentRunner = "fake-codex"
    return nil
}
```
