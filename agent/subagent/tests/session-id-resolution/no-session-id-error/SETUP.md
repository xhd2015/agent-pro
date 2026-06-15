## Preconditions
- No `--session-id` flag, no env var, no `CODEX_THREAD_ID`.
- The resolution should fail with an error that includes a generated session ID.

## Steps
1. Leave `req.SessionID` empty.
2. Ensure no session-related env vars are set.
3. Run with `Status: true` to trigger session ID resolution.
4. Verify the error message.

```go
import (
    "os"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.SessionID = ""
    os.Unsetenv("CODEX_THREAD_ID")
    os.Unsetenv("AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID")
    return nil
}
```
