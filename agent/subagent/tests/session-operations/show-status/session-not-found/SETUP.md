## Preconditions
- No session with the given session ID exists.

## Steps
1. Set `req.SessionID` to a non-existent ID.
2. Do not pre-create any sessions.
3. Call `showStatus`.
4. Verify stderr shows "session not found".

```go
import (
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.SessionID = "nonexistent_session"
    req.HomeDir = t.TempDir()
    req.SessionBase = ""
    req.PreCreateDirs = nil
    req.PreCreateMeta = nil
    return nil
}
```
