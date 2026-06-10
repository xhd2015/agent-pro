## Preconditions
- The cursor uses partition `2026-06-10` and offset `42`.

## Steps
1. Select cursor round-trip.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "cursor-valid-round-trip"; return nil }
```

