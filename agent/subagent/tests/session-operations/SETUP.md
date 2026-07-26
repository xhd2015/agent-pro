## Preconditions
- `session-operations` tests verify that `listSessions`, `showStatus`, and `traceSession` work correctly with the configurable base directory.

## Steps
1. Create session directories from `req.PreCreateDirs` with metadata and events from `req.PreCreateMeta` and `req.PreCreateEvents`.
2. If `req.PreCreatePID` is true, write a pid file with the current process PID.
3. Call `subagent.Run()` with the operation flag (`ListSessions`, `Status`, or `CatchUp`).
4. Capture stdout and stderr, return as response.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = req.RoleName
    return nil
}
```
