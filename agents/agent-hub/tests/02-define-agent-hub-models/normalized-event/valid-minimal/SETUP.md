## Preconditions
- A minimal event has `event_type` and `runner`.

## Steps
1. Select the valid minimal event scenario.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "normalized-event-valid-minimal"; return nil }
```

