## Preconditions
- The `runAgent` function accepts `ctx context.Context` as its first parameter.
- The hardcoded 30-second timeout in `runAgent` is removed; the passed context is used directly.
- `fake-codex` binary is available at `../../../../fake-codex` relative to d.DOCTEST_ROOT.
- `subagent.TestExported_runAgent` wraps the unexported `runAgent`.

## Steps
1. `Setup` chain configures the `Request` with agent runner, prompt, and cancellation flag.
2. `Run` resolves the `fake-codex` path, sets env vars, creates a context (optionally canceled), and calls `subagent.TestExported_runAgent`.
3. `Assert` verifies the output/error matches expectations.

## Context
- `req.AgentRunner`: agent runner ID to use (e.g. `"fake-codex"`).
- `req.Prompt`: prompt text to send.
- `req.CancelCtx`: if true, the context passed to `runAgent` is canceled before the call.
- `req.SessionID`: session ID string (may be empty).

```go
import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "testing"

    "github.com/xhd2015/agent-pro/agent/subagent"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = req.AgentRunner
    return nil
}
```
