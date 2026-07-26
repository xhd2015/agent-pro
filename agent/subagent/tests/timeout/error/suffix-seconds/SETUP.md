## Preconditions
- Input is `"30s"` — 30 seconds with explicit suffix, which is below the 1m minimum.

## Steps
1. Set `req.Input` to `"30s"`.
2. `ParseTimeoutDuration` should return error.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Input = "30s"
    return nil
}
```
