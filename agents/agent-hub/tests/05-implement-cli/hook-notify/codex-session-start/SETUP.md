## Preconditions
- Codex SessionStart payload is provided on stdin.
## Steps
1. Notify hook.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "hook-notify-codex-session-start"; return nil }
```

