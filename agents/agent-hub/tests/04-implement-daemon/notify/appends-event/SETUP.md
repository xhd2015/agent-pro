## Preconditions
- Daemon is running.

## Steps
1. Notify a valid event.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "notify-appends-event"; return nil }
```

