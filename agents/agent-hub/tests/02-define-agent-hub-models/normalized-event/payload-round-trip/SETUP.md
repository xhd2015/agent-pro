## Preconditions
- The event contains arbitrary JSON payload.

## Steps
1. Select the payload round-trip scenario.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "normalized-event-payload-round-trip"; return nil }
```

