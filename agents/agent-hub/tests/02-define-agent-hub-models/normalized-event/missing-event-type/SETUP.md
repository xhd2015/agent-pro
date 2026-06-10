## Preconditions
- The event omits `event_type`.

## Steps
1. Select the missing event type scenario.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "normalized-event-missing-event-type"; return nil }
```

