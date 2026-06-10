## Preconditions
- No event exists in the partition.

## Steps
1. Append one event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "append-first-event-offset-zero"; return nil }
```

