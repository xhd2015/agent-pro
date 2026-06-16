## Preconditions
- Inputs that are bare numbers (no letter suffix), treated as seconds.
- Example: `"30"` → 30 seconds. Since 30s < 1m, error is expected.

## Steps
1. Leaves set `req.Input` to bare number strings.
2. Assert: `ParseTimeoutDuration` interprets as seconds, applies validation.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    // Grouping node: bare-number inputs — leaves set req.Input
    _ = t
    return nil
}
```
