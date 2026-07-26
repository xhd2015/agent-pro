## Preconditions
- Input is `"10m"` — exactly 10 minutes, the threshold where warnings stop.
- 10m ≥ 10m, so no warning to stderr.

## Steps
1. Set `req.Input` to `"10m"`.
2. `ParseTimeoutDuration` returns 10m, no error, no stderr warning.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Input = "10m"
    return nil
}
```
