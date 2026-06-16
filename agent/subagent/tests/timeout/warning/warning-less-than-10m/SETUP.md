## Preconditions
- Input is `"3m"` — 3 minutes, which is ≥ 1m but < 10m.
- No error, but a warning to stderr is expected.

## Steps
1. Set `req.Input` to `"3m"`.
2. `ParseTimeoutDuration` returns 3m, no error, warning on stderr.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.Input = "3m"
    return nil
}
```
