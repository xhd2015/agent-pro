## Preconditions
- `session-base` tests verify that `sessionsBase` (via the public `Run()` API with `ListSessions`) respects `Options.SessionBase`, the default `~/.agent-pro/subagent/<role>/sessions/`, and the `AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME` env var override.

## Steps
1. Create session directories under the target base directory.
2. Set `req.HomeDir` to a temp directory so `~` resolves predictably.
3. Set env vars from `req.Env`.
4. Call `subagent.Run()` with `ListSessions: true`.
5. Capture stdout and return.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Operation = "session_base"
    req.RoleName = "testrole"
    return nil
}
```
