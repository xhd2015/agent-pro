## Preconditions
- Input is `"1h"` — 1 hour, which is ≥ 10m, so no warning.

## Steps
1. Set `req.Input` to `"1h"`.
2. `ParseTimeoutDuration` returns 1h, no error, no stderr warning.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.Input = "1h"
    return nil
}
```
