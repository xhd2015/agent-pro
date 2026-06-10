## Preconditions
- Daemon status is requested as JSON.
## Steps
1. Start and check status.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "daemon-status-json"; return nil }
```

