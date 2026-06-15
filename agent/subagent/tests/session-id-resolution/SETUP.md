## Preconditions
- `session-id-resolution` tests verify that `resolveSessionID` (via `Run()` with `Status`) respects `AGENT_PRO_SUBAGENT_<ROLE>_SESSION_ID` env var, `--session-id` flag, and `CODEX_THREAD_ID` fallback.

## Steps
1. Set env vars from `req.Env`.
2. Call `subagent.Run()` with `Status: true` and `req.SessionID`.
3. Capture stdout and stderr.
4. Return the captured output and error.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Operation = "session_id_resolution"
    req.RoleName = "testrole"
    return nil
}
```
