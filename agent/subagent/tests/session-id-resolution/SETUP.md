## Preconditions
- `session-id-resolution` tests verify that `resolveSessionID` (via `Run()` with `Status`) respects `AGENT_PRO_SUBAGENT_<ROLE>_SESSION_ID` env var, `--session-id` flag, and `CODEX_THREAD_ID` fallback.

## Steps
1. Set env vars from `req.Env`.
2. Call `subagent.Run()` with `Status: true` and `req.SessionID`.
3. Capture stdout and stderr.
4. Return the captured output and error.

```go
import (
    "bytes"
    "fmt"
    "os"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.Operation = "session_id_resolution"
    req.RoleName = "testrole"
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    for _, e := range req.Env {
        parts := splitEnv(e)
        if len(parts) == 2 {
            os.Setenv(parts[0], parts[1])
        }
    }

    oldOut := os.Stdout
    rOut, wOut, _ := os.Pipe()
    os.Stdout = wOut

    oldErr := os.Stderr
    rErr, wErr, _ := os.Pipe()
    os.Stderr = wErr

    runErr := subagent.Run(subagent.Config{
        RoleName: req.RoleName,
    }, subagent.Options{
        Status:    true,
        SessionID: req.SessionID,
    })

    wOut.Close()
    wErr.Close()
    os.Stdout = oldOut
    os.Stderr = oldErr

    var bufOut bytes.Buffer
    bufOut.ReadFrom(rOut)

    var bufErr bytes.Buffer
    bufErr.ReadFrom(rErr)

    return &Response{Stdout: bufOut.String(), Stderr: bufErr.String(), Err: runErr}, nil
}

func splitEnv(e string) []string {
    for i := 0; i < len(e); i++ {
        if e[i] == '=' {
            return []string{e[:i], e[i+1:]}
        }
    }
    return []string{e}
}
```
