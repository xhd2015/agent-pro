## Preconditions
- Inputs that cause `ParseTimeoutDuration` to return an error.
- Includes: duration < 1m (bare number or with suffix), and invalid/unparseable strings.

## Steps
1. Leaves set `req.Input` to values that should produce errors.
2. Assert: error is not nil, with appropriate message details.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    // Grouping node: error inputs — leaves set req.Input
    _ = t
    return nil
}
```
