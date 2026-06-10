## Preconditions
- The event occurred on one day but is received on another day.

## Steps
1. Append with received time on the later day.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "append-received-at-selects-partition"; return nil }
```

