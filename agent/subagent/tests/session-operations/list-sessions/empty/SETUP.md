## Preconditions
- No session directories exist under the configured base.

## Steps
1. Use a temp home dir with no pre-created sessions.
2. Call `ListSessions`.
3. Verify the output indicates no sessions found.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.HomeDir = t.TempDir()
    req.SessionBase = ""
    req.PreCreateDirs = nil
    req.PreCreateMeta = nil
    return nil
}
```
