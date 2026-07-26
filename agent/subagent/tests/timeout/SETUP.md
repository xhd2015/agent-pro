## Preconditions
- The `timeout` tests verify the exported function `ParseTimeoutDuration` from the `subagent` package.
- The function parses a raw string from the `--timeout` flag and returns a `time.Duration`.
- Empty input defaults to `1h`. Bare numbers (no letters) are treated as seconds.
- Validations: duration < 1m → error; 1m ≤ duration < 10m → warning to stderr.

## Steps
1. Grouping nodes organize leaves by outcome category: `valid/`, `bare-number/`, `warning/`, `error/`.
2. Leaf `SETUP.md` files set `req.Input` to the raw timeout string under test.
3. `Run` captures stderr, calls `subagent.ParseTimeoutDuration(req.Input)`, restores stderr.
4. Leaf `ASSERT.md` files validate `resp.Duration`, `resp.Err`, and `resp.Stderr`.

## Context
- `Req.Input`: raw string from `--timeout` flag (may be empty)
- `Resp.Duration`: parsed `time.Duration` (int64 nanoseconds)
- `Resp.Stderr`: captured stderr output (warnings)
- `Resp.Err`: error returned by `ParseTimeoutDuration`, if any

```go
import (
    "bytes"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/xhd2015/agent-pro/agent/subagent"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    // Root Setup: grouping nodes and leaves provide their own Setup
    _ = t
    _ = req
    return nil
}
```
