## Preconditions
- One event already exists in the partition.

## Steps
1. Append a second event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "append-second-event-offset-one"; return nil }
```

