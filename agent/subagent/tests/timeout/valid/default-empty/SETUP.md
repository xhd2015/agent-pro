## Preconditions
- Input is empty string "" — the default case.

## Steps
1. Set `req.Input` to `""`.
2. `ParseTimeoutDuration` should return `1h` with no error and no stderr warning.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Input = ""
    return nil
}
```
