# Scenario

**Feature**: Require policy errors when no session ID source is available

```
# default AutoGenerateSessionID=false
resolveSessionID (all sources miss) -> error with retry hint (gen_*, --session-id)
```

## Preconditions
- No `--session-id` flag, no env var, no `CODEX_THREAD_ID`.
- The resolution should fail with an error that includes a generated session ID.
- Env force-unset is handled by harness `envLookup` (no process Unsetenv).

## Steps
1. Leave `req.SessionID` empty.
2. Ensure no session-related env vars via harness EnvLookup force-unset.
3. Run with `Status: true` to trigger session ID resolution.
4. Verify the error message.

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.SessionID = ""
    req.AutoGenerateSessionID = false
    return nil
}
```
