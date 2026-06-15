## Preconditions
- `Config.SessionEnvVar` is empty (default behavior).
- The default env var `AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID` is set.

## Steps
1. Leave `req.SessionEnvVar` empty.
2. Set `AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID=my_sid_default` in `req.Env`.
3. Call `Run()` with `Status: true`.
4. Verify stderr shows the session was looked up using `my_sid_default`.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.SessionEnvVar = ""
    req.SessionID = ""
    req.RoleName = "testrole"
    req.Env = append(req.Env,
        "AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID=my_sid_default",
    )
    return nil
}
```
