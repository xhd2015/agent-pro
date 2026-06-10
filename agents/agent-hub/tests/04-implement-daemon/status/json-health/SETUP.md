## Preconditions
- Daemon is running.

## Steps
1. Read status.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "status-json-health"; return nil }
```

