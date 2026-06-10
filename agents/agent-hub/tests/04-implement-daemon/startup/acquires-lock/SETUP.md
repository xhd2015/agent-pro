## Preconditions
- Daemon starts with an empty home.

## Steps
1. Start daemon and inspect lock.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "startup-acquires-lock"; return nil }
```

