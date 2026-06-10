## Preconditions
- Daemon is initially stopped.
## Steps
1. Start, status, and stop.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "daemon-start-status-stop"; return nil }
```

