## Preconditions
- Input is `"30"` — a bare number with no suffix, treated as seconds.
- 30 seconds is less than 1 minute, so an error is expected.

## Steps
1. Set `req.Input` to `"30"`.
2. `ParseTimeoutDuration` should interpret it as 30s and return an error (< 1m).

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Input = "30"
    return nil
}
```
