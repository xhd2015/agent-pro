## Preconditions
- Inputs that parse successfully but fall in the warning range: 1m ≤ duration < 10m.
- A warning message is printed to stderr suggesting a longer timeout (e.g., `--timeout=1h`).

## Steps
1. Leaves set `req.Input` to values between 1m and 10m (inclusive lower, exclusive upper).
2. Assert: no error, warning on stderr, correct duration.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    // Grouping node: warning-range inputs — leaves set req.Input
    _ = t
    return nil
}
```
