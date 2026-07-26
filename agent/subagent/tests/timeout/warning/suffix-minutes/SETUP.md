## Preconditions
- Input is `"5m"` — 5 minutes, which is ≥ 1m but < 10m.
- A warning to stderr is expected.

## Steps
1. Set `req.Input` to `"5m"`.
2. `ParseTimeoutDuration` returns 5m, no error, warning on stderr.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Input = "5m"
    return nil
}
```
