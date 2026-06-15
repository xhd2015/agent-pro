## Preconditions
- Tests for `Config.SessionMetaField`: when set, `sessionMatchField()` returns the custom field name, and sessions are matched using that field in meta.json.

## Steps
1. Set `req.Status = true` to exercise session matching via `showStatus`.
2. Each leaf creates a session with a meta.json under the custom or default field name.
3. Each leaf sets env var + SessionMetaField so the session is found or not found accordingly.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.Operation = "session_meta_field"
    req.Status = true
    req.RoleName = "testrole"
    return nil
}
```
