## Preconditions
- opencode session.created payload is provided on stdin.
## Steps
1. Notify hook.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "hook-notify-opencode-session-created"; return nil }
```

