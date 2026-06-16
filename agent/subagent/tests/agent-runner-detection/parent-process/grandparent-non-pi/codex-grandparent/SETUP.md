## Preconditions
- `ppid` process name is `"bash"`.
- `pppid` process name is `"codex"` (a non-pi agent at grandparent level).

## Steps
1. Set `req.ProcessNames` to `["bash", "codex"]`.
2. ppid no match → check pppid for "pi" only → "codex" does not match.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ProcessNames = []string{"bash", "codex"}
    return nil
}
```
