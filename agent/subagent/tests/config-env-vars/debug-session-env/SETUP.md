## Preconditions
- Tests for `Config.DebugSessionEnv`: when set, `sessionsBase()` reads from that env var instead of the default `AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME`.

## Steps
1. Set `req.ListSessions = true` to exercise sessionsBase resolution.
2. Each leaf sets `req.DebugSessionEnv` to custom or empty and provides sessions at the corresponding directory.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.Operation = "debug_session_env"
    req.ListSessions = true
    req.RoleName = "testrole"
    return nil
}
```
