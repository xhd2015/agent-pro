## Preconditions
- Input is `"abc"` — a completely invalid duration string.
- `time.ParseDuration` will fail on it.

## Steps
1. Set `req.Input` to `"abc"`.
2. `ParseTimeoutDuration` returns error (parse failure).

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Input = "abc"
    return nil
}
```
