## Preconditions
- `Config.SessionEnvVar` is set to `"MY_CUSTOM_SESSION_ID"`.
- The env var `MY_CUSTOM_SESSION_ID` is set to a known value.

## Steps
1. Set `req.SessionEnvVar = "MY_CUSTOM_SESSION_ID"`.
2. Set `MY_CUSTOM_SESSION_ID=my_sid_custom` in `req.Env`.
3. Do NOT set the default `AGENT_PRO_SUBAGENT_*` env var.
4. Call `Run()` with `Status: true`.
5. Verify stderr shows the session was looked up using `my_sid_custom`.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.SessionEnvVar = "MY_CUSTOM_SESSION_ID"
    req.SessionID = ""
    req.RoleName = "testrole"
    req.Env = append(req.Env,
        "MY_CUSTOM_SESSION_ID=my_sid_custom",
    )
    return nil
}
```
