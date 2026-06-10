## Preconditions
- Daemon is running.

## Steps
1. Stop daemon.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "shutdown-releases-lock"; return nil }
```

