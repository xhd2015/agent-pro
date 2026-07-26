# Scenario

**Feature**: --session-id flag overrides env var

```
# flag has highest priority
--session-id flag_session_456 + env var -> resolveSessionID uses flag
```

## Preconditions
- Both `AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID` env var and `--session-id` flag are set.
- The flag should take priority.

## Steps
1. Set `AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID=env_session_123` in `req.Env`.
2. Set `req.SessionID = "flag_session_456"`.
3. Verify the flag value is used for session lookup.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.SessionID = "flag_session_456"
    req.Env = append(req.Env,
        "AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID=env_session_123",
    )
    return nil
}
```
