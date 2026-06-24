# Scenario

**Feature**: auto-generate session ID when all resolution sources miss

```
# AutoGenerateSessionID=true bypasses Require error path
resolveSessionID (no flag, no env, no CODEX_THREAD_ID) -> generateSessionID -> showStatus

# generated ID has no session dir yet
showStatus -> findSession -> stderr "session not found"
```

## Preconditions
- No `--session-id` flag, no `AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID`, no `CODEX_THREAD_ID`.
- `Config.AutoGenerateSessionID` is `true`.

## Steps
1. Leave `req.SessionID` empty.
2. Unset `CODEX_THREAD_ID` and `AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID`.
3. Set `req.AutoGenerateSessionID = true`.
4. Run with `Status: true` to trigger session ID resolution and lookup.

```go
import (
    "os"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.SessionID = ""
    req.AutoGenerateSessionID = true
    os.Unsetenv("CODEX_THREAD_ID")
    os.Unsetenv("AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID")
    return nil
}
```