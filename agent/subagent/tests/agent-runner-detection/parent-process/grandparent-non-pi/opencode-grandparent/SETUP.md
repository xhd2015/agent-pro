## Preconditions
- `ppid` process name is `"bash"`.
- `pppid` process name is `"opencode"` (a non-pi agent at grandparent level).

## Steps
1. Set `req.ProcessNames` to `["bash", "opencode"]`.
2. ppid no match → check pppid for "pi" only → "opencode" does not match.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.ProcessNames = []string{"bash", "opencode"}
    return nil
}
```
