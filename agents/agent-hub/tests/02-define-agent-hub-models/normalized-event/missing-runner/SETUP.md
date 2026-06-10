## Preconditions
- The event omits `runner`.

## Steps
1. Select the missing runner scenario.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "normalized-event-missing-runner"; return nil }
```

