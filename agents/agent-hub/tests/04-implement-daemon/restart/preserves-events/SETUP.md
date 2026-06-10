## Preconditions
- An event was acknowledged before restart.

## Steps
1. Restart daemon and fetch event.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "restart-preserves-events"; return nil }
```

