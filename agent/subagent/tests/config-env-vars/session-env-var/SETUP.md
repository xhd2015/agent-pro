## Preconditions
- Tests for `Config.SessionEnvVar`: when set, `resolveSessionID` reads from that env var instead of the default.

## Steps
1. Set `req.Status = true` to exercise session ID resolution via `showStatus`.
2. Each leaf sets `req.SessionEnvVar` to custom or empty and provides the corresponding env var.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.Operation = "session_env_var"
    req.Status = true
    return nil
}
```
