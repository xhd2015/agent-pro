## Preconditions
- `AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID` env var is set to a known session ID.
- No `--session-id` flag is provided.
- No `CODEX_THREAD_ID` env var.

## Steps
1. Set `AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID=env_session_123` in `req.Env`.
2. Leave `req.SessionID` empty (no flag).
3. Unset `CODEX_THREAD_ID`.
4. Verify the session ID resolution uses the env var.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.SessionID = ""
    req.Env = append(req.Env,
        "AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID=env_session_123",
    )
    return nil
}
```
