## Preconditions
- `ppid` process name is `"bash"`.
- `pppid` process name is `"bash"` (no agent at any level).

## Steps
1. Set `req.ProcessNames` to `["bash", "bash"]`.
2. No match at ppid or pppid → not detected.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.ProcessNames = []string{"bash", "bash"}
    return nil
}
```
