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

type Request struct {
    Input string
}

type Response struct {
    Duration time.Duration
    Stderr   string
    Err      error
}

func Setup(t *testing.T, req *Request) error {
    // Root Setup: grouping nodes and leaves provide their own Setup
    _ = t
    _ = req
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    oldErr := os.Stderr
    rErr, wErr, err := os.Pipe()
    if err != nil {
        return nil, fmt.Errorf("create pipe: %w", err)
    }
    os.Stderr = wErr

    d, parseErr := subagent.ParseTimeoutDuration(req.Input)

    wErr.Close()
    os.Stderr = oldErr

    var bufErr bytes.Buffer
    bufErr.ReadFrom(rErr)

    return &Response{
        Duration: d,
        Stderr:   bufErr.String(),
        Err:      parseErr,
    }, nil
}
```
