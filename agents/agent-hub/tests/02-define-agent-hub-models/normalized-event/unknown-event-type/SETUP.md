## Preconditions
- The event uses an unsupported event type.

## Steps
1. Select the unknown event type scenario.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "normalized-event-unknown-event-type"; return nil }
```

