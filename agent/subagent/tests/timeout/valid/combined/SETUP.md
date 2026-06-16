## Preconditions
- Input is `"1h30m"` — a combined duration, which is ≥ 10m, so no warning.

## Steps
1. Set `req.Input` to `"1h30m"`.
2. `ParseTimeoutDuration` returns 1h30m, no error, no stderr warning.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.Input = "1h30m"
    return nil
}
```
