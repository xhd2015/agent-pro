## Preconditions
- A previous day has an event.

## Steps
1. Append into a new day.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "append-new-day-resets-offset"; return nil }
```

