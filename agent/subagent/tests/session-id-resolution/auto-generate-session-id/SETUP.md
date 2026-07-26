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
- Env force-unset is handled by harness `envLookup` (no process Unsetenv).

## Steps
1. Leave `req.SessionID` empty.
2. Set `req.AutoGenerateSessionID = true` (harness EnvLookup force-unsets role/codex env).
3. Run with `Status: true` to trigger session ID resolution and lookup.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.SessionID = ""
    req.AutoGenerateSessionID = true
    return nil
}
```
