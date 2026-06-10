## Preconditions
- Event is missing event type.

## Steps
1. Notify invalid event.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "notify-invalid-event-rejected"; return nil }
```

