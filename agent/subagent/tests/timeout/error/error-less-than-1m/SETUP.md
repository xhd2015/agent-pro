## Preconditions
- Input is `"30s"` — 30 seconds with explicit suffix, which is below the 1m minimum.
- The error message must mention the minimum duration and suggest a longer timeout.

## Steps
1. Set `req.Input` to `"30s"`.
2. `ParseTimeoutDuration` returns an error containing "at least 1m" or similar.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Input = "30s"
    return nil
}
```
