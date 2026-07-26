## Preconditions
- Inputs that parse successfully with no error and no warning.
- Includes: default empty, explicit durations ≥ 10m, and the 10m boundary.

## Steps
1. Leaves set `req.Input` to values that should parse cleanly.
2. Assert: no error, no stderr warning, correct duration.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    // Grouping node: valid inputs — leaves set req.Input
    _ = t
    return nil
}
```
