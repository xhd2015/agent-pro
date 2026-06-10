## Preconditions
- Daemon starts with an empty home.

## Steps
1. Start daemon and inspect socket.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "startup-creates-socket"; return nil }
```

